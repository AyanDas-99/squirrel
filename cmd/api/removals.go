package main

import (
	"errors"
	"net/http"

	"test.com/internal/data"
	"test.com/internal/validator"
)

func (app *application) addRemoval(w http.ResponseWriter, r *http.Request) {
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

	validator := validator.New()
	validator.Check(input.ItemID != 0, "item_id", "must be provided")
	validator.Check(input.Quantity != 0, "quantity", "must be provided")
	validator.Check(input.Quantity > 0, "quantity", "must be greater than 0")
	validator.Check(input.Remarks != "", "remarks", "must be provided")

	if !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	removal := &data.Removal{
		ItemID:   input.ItemID,
		Quantity: input.Quantity,
		Remarks:  input.Remarks,
	}

	item, err := app.items.GetItem(removal.ItemID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if item.Remaining < removal.Quantity {
		app.failedValidationResponse(w, r, map[string]string{"item": "item is not available in the required quantity"})
		return
	}

	tx, err := app.removals.DB.Begin()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer tx.Rollback()

	err = app.removals.InsertRemoval(tx, removal)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.items.UpdateRemaining(tx, removal.ItemID, removal.Quantity, item.Version)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"removal": removal}, nil)
}

func (app *application) listRemovals(w http.ResponseWriter, r *http.Request) {
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
	input.Filters.Sort = app.readString(qs, "sort", "removed_at")
	input.Filters.SortSafelist = []string{"id", "-id", "removed_at", "-removed_at"}

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

	removals, metadata, err := app.removals.GetRemovals(input.ItemID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"removals": removals, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
