package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/hwanbin/wanpm-api/internal/data"
	"github.com/hwanbin/wanpm-api/internal/validator"
	"github.com/oklog/ulid/v2"
)

func (app *application) createTimesheetHandler(w http.ResponseWriter, r *http.Request) {
	var input data.TimesheetInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Description == nil {
		*input.Description = ""
	}

	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateTimesheetInputRequired(v, &input); !v.Valid() {
		app.missingRequiredFieldsResponse(w, r, v.Errors)
		return
	}
	// if data.ValidateTimesheetInputSemantic(v, &input); !v.Valid() {
	// 	app.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }
	timesheetRequest := &data.Timesheet{
		ID:          id.String(),
		UserID:      *input.UserID,
		ProjectID:   *input.ProjectID,
		ClientID:    *input.ClientID,
		ActivityID:  *input.ActivityID,
		WorkDate:    *input.WorkDate,
		WorkMinutes: *input.WorkMinutes,
		Description: *input.Description,
		Status:      data.StatusActive,
	}
	if data.ValidateTimesheetSemantic(v, timesheetRequest); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// go routine?
	_, err = app.models.Project.Get(*input.ProjectID)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("project id: %v", err.Error()))
		return
	}
	_, err = app.models.Client.Get(*input.ClientID)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("client id: %v", err.Error()))
		return
	}
	_, err = app.models.User.GetById(*input.UserID)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("user id: %v", err.Error()))
		return
	}
	_, err = app.models.Activity.Get(*input.ActivityID)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("activity id: %v", err.Error()))
		return
	}

	// check UserID is associated with the input projectID
	assignments, err := app.models.Assignment.GetAllByProjectID(*input.ProjectID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if data.ValidateUserID(v, assignments, *input.UserID); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// check ClientID is associated with the input projectID
	associatedClients, err := app.models.Client.GetAllByProjectID(*input.ProjectID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if data.ValidateClientID(v, associatedClients, *input.ClientID); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// timesheetRequest := &data.Timesheet{
	// 	ID:          id.String(),
	// 	UserID:      *input.UserID,
	// 	ProjectID:   *input.ProjectID,
	// 	ClientID:    *input.ClientID,
	// 	ActivityID:  *input.ActivityID,
	// 	WorkDate:    *input.WorkDate,
	// 	WorkMinutes: *input.WorkMinutes,
	// 	Description: *input.Description,
	// 	Status:      data.StatusActive,
	// }
	// // if input.Description != nil {
	// // 	timesheetRequest.Description = *input.Description
	// // }
	// if data.ValidateTimesheetSemantic(v, timesheetRequest); !v.Valid(){
	// 	app.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

	err = app.models.Timesheet.Insert(timesheetRequest)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateTimesheetID):
			app.serverErrorResponse(w, r, err)
			return
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/timesheet/%s", timesheetRequest.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"timesheeet": timesheetRequest}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showTimesheetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readStringIDParam(r)
	if err != nil {
		// this handler won't be called if there's no id path param
		app.badRequestResponse(w, r, err)
		return
	}
	if !data.UlidRegex.MatchString(id) {
		app.badRequestResponse(w, r, errors.New("invalid id"))
		return
	}

	timesheet, err := app.models.Timesheet.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"timesheet": timesheet}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listTimesheetHandler(w http.ResponseWriter, r *http.Request) {
	var input data.TimesheetQsInput

	qs := r.URL.Query()
	v := validator.New()

	input.UserID = app.readString(qs, "user_id", "")
	input.Email = app.readString(qs, "email", "")
	input.FirstName = app.readString(qs, "first_name", "")
	input.LastName = app.readString(qs, "last_name", "")
	input.ProjectID = int32(app.readInt(qs, "project_id", 0, v))
	input.ProjectName = app.readString(qs, "project_name", "")
	input.ActivityID = int32(app.readInt(qs, "activity_id", 0, v))
	input.ActivityName = app.readString(qs, "activity_name", "")
	input.WorkDate = app.readString(qs, "work_date", "")
	input.FromWorkDate = app.readString(qs, "from_date", "")
	input.ToWorkDate = app.readString(qs, "to_date", "")

	// TODO: validate query string
	// Must - ToWorkDate >= FromWorkDate
	// range operation needs both FromWorkDate and ToWorkDate pair

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 0, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "-id", "project_id", "-project_id", "work_date", "-work_date"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	timesheets, metaData, err := app.models.Timesheet.GetAll(input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metaData, "timesheets": timesheets}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateTimesheetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readStringIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	timesheet, err := app.models.Timesheet.InternalGet(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input data.TimesheetInput
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.UserID != nil {
		timesheet.UserID = *input.UserID
	}
	if input.ProjectID != nil {
		timesheet.ProjectID = *input.ProjectID
	}
	if input.ClientID != nil {
		timesheet.ClientID = *input.ClientID
	}
	if input.ActivityID != nil {
		timesheet.ActivityID = *input.ActivityID
	}
	if input.WorkDate != nil {
		timesheet.WorkDate = *input.WorkDate
	}
	if input.WorkMinutes != nil {
		timesheet.WorkMinutes = *input.WorkMinutes
	}
	if input.Description != nil {
		timesheet.Description = *input.Description
	}

	v := validator.New()
	if data.ValidateTimesheetSemantic(v, timesheet); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// check UserID is associated with the input projectID
	assignments, err := app.models.Assignment.GetAllByProjectID(timesheet.ProjectID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if data.ValidateUserID(v, assignments, timesheet.UserID); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// check ClientID is associated with the input projectID
	associatedClients, err := app.models.Client.GetAllByProjectID(timesheet.ProjectID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if data.ValidateClientID(v, associatedClients, timesheet.ClientID); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Timesheet.Update(timesheet)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	joinedTimesheet, err := app.models.Timesheet.Get(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"timesheet": joinedTimesheet}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteTimesheetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readStringIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Timesheet.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "timesheet successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
