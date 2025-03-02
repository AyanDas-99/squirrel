package main

import (
	"errors"
	"net/http"

	"test.com/internal/data"
	"test.com/internal/validator"
)

func (app *application) addIssue(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID   int64  `json:"item_id"`
		Quantity int32  `json:"quantity"`
		IssuedTo string `json:"issued_to"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	validator := validator.New()
	validator.Check(input.ItemID != 0, "item_id", "Field cannot be blank")
	validator.Check(input.ItemID > 0, "item_id", "Field cannot be negative")
	validator.Check(input.Quantity > 0, "quantity", "Field must be positive integer")
	validator.Check(input.IssuedTo != "", "issued_to", "Field cannot be blank")

	if !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	issue := &data.Issue{
		ItemID:   input.ItemID,
		Quantity: input.Quantity,
		IssuedTo: input.IssuedTo,
	}

	item, err := app.items.GetItem(issue.ItemID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if item.Remaining == 0 {
		app.failedValidationResponse(w, r, map[string]string{"item": "item is not available"})
		return
	} else if item.Remaining < issue.Quantity {
		app.failedValidationResponse(w, r, map[string]string{"item": "item is not available in the required quantity"})
		return
	}

	// Begin transaction to issue and update item
	tx, err := app.items.DB.Begin()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer tx.Rollback()

	err = app.issues.InsertIssue(tx, issue)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.items.UpdateRemaining(tx, issue.ItemID, issue.Quantity, item.Version)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = tx.Commit()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"issue": issue}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Always sorted by issued_at
func (app *application) listIssues(w http.ResponseWriter, r *http.Request) {

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
	input.Filters.Sort = app.readString(qs, "sort", "-issued_at")
	input.Filters.SortSafelist = []string{"id", "-id", "issued_at", "-issued_at"}

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

	issues, metadata, err := app.issues.GetIssues(input.ItemID, input.Filters)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"issues": issues, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
