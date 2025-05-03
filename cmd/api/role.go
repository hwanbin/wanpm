package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hwanbin/wanpm/internal/data"
	"github.com/hwanbin/wanpm/internal/validator"
)

func (app *application) createRoleHandler(w http.ResponseWriter, r *http.Request) {
	var input data.RoleInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateRoleInputRequired(v, &input); !v.Valid() {
		app.missingRequiredFieldsResponse(w, r, v.Errors)
		return
	}
	if data.ValidateRoleInputSemantic(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	roleRequest := &data.Role{
		Name: *input.Name,
	}

	err = app.models.Role.Insert(roleRequest)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateRoleName):
			v.AddError("name", "role with this name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/role/%d", roleRequest.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"role": roleRequest}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	role, err := app.models.Role.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"role": role}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listRoleHandler(w http.ResponseWriter, r *http.Request) {
	var input data.RoleQsInput

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

	roles, metaData, err := app.models.Role.GetAll(input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metaData, "roles": roles}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	roleRequest, err := app.models.Role.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input data.RoleInput
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateRoleInputRequired(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	if data.ValidateRoleInputSemantic(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	roleRequest.Name = *input.Name

	err = app.models.Role.Update(roleRequest)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	roleResponse, err := app.models.Role.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"role": roleResponse}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readInt32IDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Role.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "role successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
