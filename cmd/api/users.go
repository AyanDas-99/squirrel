package main

import (
	"errors"
	"net/http"

	"test.com/internal/data"
	"test.com/internal/validator"
)

func (app *application) registerUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.UserName != "", "username", "must be provided")
	v.Check(len(input.UserName) >= 3, "username", "must be at least 3 bytes long")
	v.Check(input.Password != "", "password", "must be provided")
	v.Check(len(input.Password) == 4, "password", "must be 4 bytes long")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := &data.User{
		UserName: input.UserName,
		Password: input.Password,
		IsAdmin:  false}

	hash, err := data.PasswordToHash(user.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	user.Hash = hash
	err = app.users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateName):
			v.AddError("username", "must be unique")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.users.GetAllNonAdmin()
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserPermissionById(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIdFromParams(r)
	if err != nil || id < 1 {
		app.notFoundErrorResponse(w, r)
		return
	}

	permissions, err := app.permissions.GetAllForUser(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"permissions": permissions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updatePermission(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID       int64    `json:"user_id"`
		PermissionID []string `json:"permission_ids"`
		Grant        bool     `json:"grant"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Grant {
		err = app.permissions.AddForUser(input.UserID, input.PermissionID...)
	} else {
		err = app.permissions.RemoveForUser(input.UserID, input.PermissionID...)
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "permission added"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
