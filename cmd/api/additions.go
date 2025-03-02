package main

import (
	"errors"
	"net/http"

	"test.com/internal/data"
	"test.com/internal/validator"
)

func (app *application) refillItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID   int64  `json:"item_id"`
		Quantity int32  `json:"quantity"`
		Remarks  string `json:"remarks"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.ItemID != 0, "item_id", "must be provided")
	v.Check(input.ItemID > 0, "item_id", "must be greater than 0")

	v.Check(input.Quantity != 0, "quantity", "must be provided")
	v.Check(input.Quantity > 0, "quantity", "must be greater than 0")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	item, err := app.items.GetItem(input.ItemID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	tx, err := app.additions.DB.Begin()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	addition := &data.Addition{
		ItemID:   input.ItemID,
		Quantity: input.Quantity,
		Remarks:  input.Remarks,
	}

	err = app.additions.InsertAddition(tx, addition)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.items.AddRemaining(tx, input.ItemID, input.Quantity, item.Version)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"addition": addition}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) listRefills(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdFromParams(r)
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}

	v := validator.New()

	qs := r.URL.Query()
	var input struct {
		ItemID int64
		data.Filters
	}
	input.ItemID = id
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 10, v)
	input.Filters.Sort = app.readString(qs, "sort", "-added_at")
	input.Filters.SortSafelist = []string{"id", "-id", "added_at", "-added_at"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	v.Check(input.ItemID != 0, "item_id", "Field cannot be blank")
	v.Check(input.ItemID > 0, "item_id", "Field cannot be negative")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	additions, metadata, err := app.additions.GetAdditions(id, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"additions": additions, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
