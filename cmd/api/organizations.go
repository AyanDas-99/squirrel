package main

import (
	"errors"
	"net/http"

	"test.com/internal/data"
	"test.com/internal/validator"
)

func (app *application) addOrg(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.Name != "", "name", "must be provided")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	org := data.Organization{
		Name: input.Name}

	err = app.org.InsertOrganization(&org)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateName):
			v.AddError("name", "organization with name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"org": org}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listOrgs(w http.ResponseWriter, r *http.Request) {
	orgs, err := app.org.GetOrganizations()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"orgs": orgs}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getOrgForId(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdFromParams(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	org, err := app.org.GetOrganizationByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"org": org}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
