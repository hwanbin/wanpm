package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hwanbin/wanpm/internal/data"
	"github.com/hwanbin/wanpm/internal/validator"
)

func (app *application) createProposalHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ExternalID string `json:"proposal_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	proposal := &data.Proposal{
		ExternalID: input.ExternalID,
	}

	v := validator.New()
	if data.ValidateProposal(v, proposal); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Proposal.Insert(proposal)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/proposal/%s", proposal.ExternalID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"proposal": proposal}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProposalHandler(w http.ResponseWriter, r *http.Request) {
	externalID, err := app.readStringIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	proposal, err := app.models.Proposal.Get(externalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"proposal": proposal}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProposalHandler(w http.ResponseWriter, r *http.Request) {
	externalID, err := app.readStringIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	proposal, err := app.models.Proposal.Get(externalID)
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
		ExternalID string `json:"proposal_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	proposal.ExternalID = input.ExternalID

	v := validator.New()
	if data.ValidateProposal(v, proposal); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Proposal.Update(proposal)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"proposal": proposal}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProposalHandler(w http.ResponseWriter, r *http.Request) {
	externalID, err := app.readStringIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Proposal.Delete(externalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "proposal successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
