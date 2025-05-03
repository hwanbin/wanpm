package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hwanbin/wanpm/internal/data"
	"github.com/hwanbin/wanpm/internal/validator"
)

func (app *application) createActivityHandler(w http.ResponseWriter, r *http.Request) {
	var input data.ActivityInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateActivityInputRequired(v, &input); !v.Valid() {
		app.missingRequiredFieldsResponse(w, r, v.Errors)
		return
	}
	if data.ValidateActivityInputSemantic(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	activity := &data.Activity{
		Name: *input.Name,
	}

	err = app.models.Activity.Insert(activity)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateActivityName):
			v.AddError("name", "activity with this name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/activity/%d", activity.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"activity": activity}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showActivityHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	activity, err := app.models.Activity.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"activity": activity}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listActivityHandler(w http.ResponseWriter, r *http.Request) {
	var input data.ActivityQsInput

	qs := r.URL.Query()
	v := validator.New()

	input.Name = app.readString(qs, "name", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 0, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"name", "-name", "id", "-id"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	activities, metaData, err := app.models.Activity.GetAll(input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metaData, "activities": activities}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateActivityHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	activityRequest, err := app.models.Activity.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input data.ActivityInput
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateActivityInputRequired(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	if data.ValidateActivityInputSemantic(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	activityRequest.Name = *input.Name

	err = app.models.Activity.Update(activityRequest)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	activityResponse, err := app.models.Activity.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"activity": activityResponse}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteActivityHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Activity.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "activity successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
