package main

import (
	"errors"
	"net/http"

	"test.com/internal/data"
	"test.com/internal/validator"
)

func (app *application) insertTag(w http.ResponseWriter, r *http.Request) {
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

	tag := data.Tag{
		Name: input.Name}

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.tags.InsertTag(&tag)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateName):
			v.AddError("name", "tag with name already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"tag": tag}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) removeTag(w http.ResponseWriter, r *http.Request) {

	var input struct {
		ID int `json:"tag_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.ID != 0, "tag_id", "must be greater than 0")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.tags.DeleteTag(input.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := app.tags.GetTags()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"tags": tags}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) removeItemTag(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID int `json:"item_id"`
		TagID  int `json:"tag_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.ItemID > 0, "item_id", "must be a positive integer")
	v.Check(input.TagID > 0, "tag_id", "must be a positive integer")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.tags.RemoveItemTag(input.ItemID, input.TagID)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) addItemTag(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID int32 `json:"item_id"`
		TagID  int32 `json:"tag_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.ItemID > 0, "item_id", "must be a positive integer")
	v.Check(input.TagID > 0, "tag_id", "must be a positive integer")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	itemTag := data.ItemTag{}
	itemTag.ItemID = input.ItemID
	itemTag.TagID = input.TagID

	err = app.tags.InsertItemTag(&itemTag)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateItemTag):
			v.AddError("item_tag", "item already has this tag")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrItemIdDoesNotExists):
			v.AddError("item_id", "does not exist")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrTagIdDoesNotExists):
			v.AddError("tag_id", "does not exist")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"item_tag": itemTag}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listItemTags(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdFromParams(r)
	if err != nil || id < 1 {
		app.notFoundErrorResponse(w, r)
		return
	}

	tags, err := app.tags.GetTagsForItem(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"tags": tags}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
