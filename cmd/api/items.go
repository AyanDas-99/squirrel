package main

import (
	"errors"
	"net/http"
	"strconv"

	"test.com/internal/data"
	"test.com/internal/validator"

	"github.com/julienschmidt/httprouter"
)

func (app *application) addItem(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request body
	var input struct {
		Name     string `json:"name"`
		Quantity int32  `json:"quantity"`
		Remarks  string `json:"remarks"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	validator := validator.New()
	validator.Check(input.Name != "", "name", "Field cannot be blank")
	validator.Check(input.Quantity != 0, "quantity", "Field cannot be blank")
	validator.Check(input.Quantity > 0, "quantity", "Field cannot be negative")

	if !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	item := &data.Item{
		Name:     input.Name,
		Quantity: input.Quantity,
		Remarks:  input.Remarks,
	}

	tx, err := app.items.DB.BeginTx(r.Context(), nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer tx.Rollback()

	err = app.items.InsertItem(tx, item)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	addition := &data.Addition{
		ItemID:   item.ID,
		Quantity: item.Quantity,
		Remarks:  input.Remarks,
	}

	err = app.additions.InsertAddition(tx, addition)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"item": item}, http.Header{})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getItem(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	v := validator.New()

	v.Check(id != 0, "id", "Field cannot be blank")
	v.Check(id > 0, "id", "Field cannot be negative")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	item, err := app.items.GetItem(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getItems(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string
		Remarks string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Remarks = app.readString(qs, "remarks", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 10, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "remarks", "created_at", "-id", "-name", "-remarks", "-created_at"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	items, metadata, err := app.items.GetAllItems(input.Name, input.Remarks, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"items": items, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateItem(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	var input struct {
		Remaining int32 `json:"remaining"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	item, err := app.items.GetItem(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	item.Remaining = input.Remaining

	err = app.items.UpdateItem(item)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteItem(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	err = app.items.DeleteItem(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
}
