package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hwanbin/wanpm/internal/data"
	"github.com/hwanbin/wanpm/internal/validator"
	"github.com/nrednav/cuid2"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	generateCuid, err := cuid2.Init(cuid2.WithLength(8))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	userID := generateCuid()
	if !cuid2.IsCuid(userID) || len(userID) != 8 {
		app.serverErrorResponse(w, r, errors.New("error generating userID"))
		return
	}

	user := &data.User{
		ID:        userID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.User.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Token.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		mailData := map[string]any{
			"activationToken": token.Plaintext,
			"firstName":       user.FirstName,
			"lastName":        user.LastName,
		}
		fmt.Println(mailData)
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", mailData)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.User.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.User.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Token.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.User.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	tokenPair, err := app.config.auth.GenerateRSAedTokenPair(
		&JwtUser{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	log.Println(tokenPair)

	refreshCookie := app.config.auth.GetRefreshCookie(tokenPair.RefreshToken)
	http.SetCookie(w, refreshCookie)

	err = app.writeJSON(w, http.StatusAccepted, envelope{"access_token": tokenPair.Token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.config.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			_, err := jwt.ParseWithClaims(
				refreshToken,
				claims,
				func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}

					publicKey, err := readPublicKey()
					if err != nil {
						return nil, err
					}

					return publicKey, nil
				},
			)

			if err != nil {
				app.errorResponse(w, r, http.StatusUnauthorized, errors.New("unauthorized"))
				return
			}

			email, err := claims.GetSubject()
			if err != nil {
				app.errorResponse(w, r, http.StatusUnauthorized, errors.New("unknown user"))
				return
			}

			v := validator.New()

			data.ValidateEmail(v, email)
			if !v.Valid() {
				app.failedValidationResponse(w, r, v.Errors)
				return
			}

			user, err := app.models.User.GetByEmail(email)
			if err != nil {
				switch {
				case errors.Is(err, data.ErrRecordNotFound):
					app.invalidCredentialsResponse(w, r)
				default:
					app.serverErrorResponse(w, r, err)
				}
				return
			}

			tokenPair, err := app.config.auth.GenerateRSAedTokenPair(
				&JwtUser{
					Email:     user.Email,
					FirstName: user.FirstName,
					LastName:  user.LastName,
				},
			)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			http.SetCookie(w, app.config.auth.GetRefreshCookie(tokenPair.RefreshToken))

			app.writeJSON(w, http.StatusOK, envelope{"access_token": tokenPair.Token}, nil)
		}
	}
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.config.auth.GetExpiredRefreshCookie())

	w.WriteHeader(http.StatusAccepted)
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readStringIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.models.User.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	var input data.UserQsInput

	qs := r.URL.Query()
	v := validator.New()

	input.Email = app.readString(qs, "email", "")
	input.FirstName = app.readString(qs, "first_name", "")
	input.LastName = app.readString(qs, "last_name", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 0, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "-id", "email", "-email", "first_name", "-first_name", "last_name", "-last_name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	users, metaData, err := app.models.User.GetAll(input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metaData, "users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
