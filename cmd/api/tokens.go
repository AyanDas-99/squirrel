package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"test.com/internal/data"
	"test.com/internal/validator"
)

func (app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.Username != "", "username", "must be provided")
	v.Check(len(input.Username) >= 3, "username", "must be at least 3 bytes long")
	v.Check(input.Password != "", "password", "must be provided")
	v.Check(len(input.Password) == 4, "password", "must be 4 bytes long")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.users.GetUserByUserName(input.Username)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	matches, err := data.CheckPasswordOnHash(input.Password, user.Hash)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !matches {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.tokens.New(user.ID, 24*time.Hour)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) validateToken(w http.ResponseWriter, r *http.Request) {

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.users.GetForToken(token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrNoRecord):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		err = app.writeJSON(w, http.StatusOK, envelope{"user":user}, nil)
		if err != nil {

		}
}
