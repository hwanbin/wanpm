package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hwanbin/wanpm-api/internal/data"
	"github.com/hwanbin/wanpm-api/internal/s3action"
	"github.com/hwanbin/wanpm-api/internal/validator"
)

func (app *application) createProjectHandler(w http.ResponseWriter, r *http.Request) {
	var input data.ProjectInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateProjectInputRequired(v, &input); !v.Valid() {
		app.missingRequiredFieldsResponse(w, r, v.Errors)
		return
	}
	if data.ValidateProjectInputSemantic(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	feature, err := json.Marshal(input.Feature)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("failed to marshal feature json: %v", err))
		return
	}
	inputFeature := string(feature)

	project := &data.ProjectRequest{
		ExternalID: input.ExternalID,
		ProposalID: input.ProposalID,
		Name:       input.Name,
		Status:     input.Status,
		Feature:    &inputFeature,
		Images:     input.Images,
	}

	project.Clients = []data.ProjectClient{}
	for _, clientName := range input.ClientNames {
		client, err := app.models.Client.GetClientByName(clientName)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				v.AddError("client_names", fmt.Sprintf("%s cannot be found", clientName))
				app.failedValidationResponse(w, r, v.Errors)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		project.Clients = append(project.Clients, data.ProjectClient{
			ClientID:      &client.InternalID,
			ClientName:    client.Name,
			ClientLogo:    client.LogoURL,
			ClientAddress: client.Address,
			ClientNote:    client.Note,
		})
	}

	err = app.models.Project.Insert(project)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateProjectID):
			v.AddError("project_id", "a project with this project_id already exists")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrDuplicateProposalID):
			v.AddError("proposal_id", "a proposal with this proposal_id already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	projectResponse, err := app.models.Project.Get(*project.ExternalID)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unable to get the project: %v", err))
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/project/%d", *projectResponse.ExternalID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"project": projectResponse}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProjectHandler(w http.ResponseWriter, r *http.Request) {
	externalID, err := app.readInt32IDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	project, err := app.models.Project.Get(externalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"project": project}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) toProject(project *data.ProjectResponse, input *data.ProjectInput) {
	if input.ExternalID != nil {
		project.ExternalID = input.ExternalID
	}

	if input.ProposalID != nil {
		project.ProposalID = input.ProposalID
	}

	if input.Name != nil {
		project.Name = input.Name
	}

	if input.Status != nil {
		project.Status = input.Status
	}

	if input.Feature != nil {
		project.Feature = input.Feature
	}

	if input.Images != nil {
		if len(input.Images) == 0 {
			project.Images = nil
		} else {
			project.Images = input.Images
		}
	}
}

func (app *application) updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	externalID, err := app.readInt32IDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	project, err := app.models.Project.Get(externalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input data.ProjectInput
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUpdatingProjectInput(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	app.toProject(project, &input)

	feature, err := json.Marshal(project.Feature)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("failed to marshal feature json: %v", err))
		return
	}
	inputFeature := string(feature)

	projectRequest := &data.ProjectRequest{
		InternalID: project.InternalID,
		ExternalID: project.ExternalID,
		ProposalID: project.ProposalID,
		Name:       project.Name,
		Status:     project.Status,
		Feature:    &inputFeature,
		Images:     project.Images,
		Version:    project.Version,
		CreatedAt:  project.CreatedAt,
		UpdatedAt:  project.UpdatedAt,
	}

	if input.ClientNames != nil {
		projectRequest.Clients = []data.ProjectClient{}
		for _, clientName := range input.ClientNames {
			client, err := app.models.Client.GetClientByName(clientName)
			if err != nil {
				switch {
				case errors.Is(err, data.ErrRecordNotFound):
					v.AddError("client_names", fmt.Sprintf("%s cannot be found", clientName))
					app.failedValidationResponse(w, r, v.Errors)
				default:
					app.serverErrorResponse(w, r, err)
				}
				return
			}
			projectRequest.Clients = append(projectRequest.Clients, data.ProjectClient{
				ClientID:      &client.InternalID,
				ClientName:    client.Name,
				ClientLogo:    client.LogoURL,
				ClientAddress: client.Address,
				ClientNote:    client.Note,
			})
		}
	} else {
		projectRequest.Clients = project.Clients
	}

	err = app.models.Project.Update(projectRequest)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	projectResponse, err := app.models.Project.Get(*projectRequest.ExternalID)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unable to get the project: %v", err))
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"project": projectResponse}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	externalID, err := app.readInt32IDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	project, err := app.models.Project.Get(externalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var objects []types.ObjectIdentifier
	fileNames, err := s3action.ListObjects(
		app.s3actor.client,
		app.config.s3.bucket,
		strconv.Itoa(int(externalID)),
	)
	for _, fileName := range fileNames {
		objects = append(
			objects,
			types.ObjectIdentifier{
				Key: &fileName,
			},
		)
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Project.Delete(
		project.InternalID,
		app.config.s3.bucket,
		strconv.Itoa(int(externalID)),
		app.s3actor.client,
		objects,
	)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "project successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listProjectHandler(w http.ResponseWriter, r *http.Request) {
	var input data.ProjectQsInput

	qs := r.URL.Query()
	v := validator.New()

	input.Name = app.readString(qs, "name", "")
	input.Status = app.readString(qs, "status", "")
	input.ProposalId = app.readString(qs, "proposal_id", "")
	input.ProjectId = app.readString(qs, "project_id", "")
	input.FullAddress = app.readString(qs, "full_address", "")
	input.ClientName = app.readString(qs, "client_name", "")
	input.Clients = app.readCSV(qs, "clients", []string{})
	input.Bbox = app.readCSV(qs, "bbox", nil)

	if data.ValidateQueryString(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	floatBbox, err := data.ConvertToBbox(input.Bbox)
	if err != nil {
		var numErr *strconv.NumError
		switch {
		case errors.As(err, &numErr):
			v.AddError("bbox", err.Error())
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, strconv.ErrRange):
			v.AddError("bbox", "a value is out of range")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 0, v)
	input.Filters.Sort = app.readString(qs, "sort", "project_id")
	input.Filters.SortSafelist = []string{"project_id", "name", "status", "-project_id", "-name", "-status"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	projects, metadata, err := app.models.Project.GetAll(
		input,
		floatBbox,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "projects": projects}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
