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
