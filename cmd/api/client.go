package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hwanbin/wanpm/internal/data"
	"github.com/hwanbin/wanpm/internal/validator"
)

func (app *application) createClientHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    *string `json:"name"`
		Address *string `json:"address"`
		LogoURL *string `json:"logo_url"`
		Note    *string `json:"note"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	client := &data.Client{
		Name:    input.Name,
		Address: input.Address,
		LogoURL: input.LogoURL,
		Note:    input.Note,
	}

	v := validator.New()
	if data.ValidateClient(v, client); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Client.Insert(client)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/client/%d", client.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"client": client}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showClientHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	client, err := app.models.Client.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"client": client}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listClientHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 0, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "-id", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	clients, metadata, err := app.models.Client.GetAll(input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "clients": clients}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateClientHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	client, err := app.models.Client.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name    *string `json:"name"`
		Address *string `json:"address"`
		Note    *string `json:"note"`
		LogoURL *string `json:"logo_url"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		client.Name = input.Name
	}

	if input.Address != nil {
		client.Address = input.Address
	}

	if input.Note != nil {
		client.Note = input.Note
	}

	if input.LogoURL != nil {
		client.LogoURL = input.LogoURL
	}

	v := validator.New()

	if data.ValidateClient(v, client); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Client.Update(client)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"client": client}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteClientHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Client.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "client successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
