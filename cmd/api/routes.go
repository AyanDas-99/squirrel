package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Create a new router
	router := httprouter.New()

	// Define routes
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprintf(w, "Welcome to sqirrel")
	})

	// Handle 404
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/items", app.requireAuthenticatedUser(app.getItems))
	router.HandlerFunc(http.MethodGet, "/items/:id", app.getItem)
	router.HandlerFunc(http.MethodPost, "/items", app.addItem)
	router.HandlerFunc(http.MethodPut, "/items/:id", app.updateItem)
	router.HandlerFunc(http.MethodDelete, "/items/:id", app.deleteItem)
	router.HandlerFunc(http.MethodGet, "/issues/:id", app.listIssues)
	router.HandlerFunc(http.MethodPost, "/issues", app.addIssue)
	router.HandlerFunc(http.MethodPost, "/removals", app.addRemoval)
	router.HandlerFunc(http.MethodGet, "/removals/:id", app.listRemovals)
	router.HandlerFunc(http.MethodPost, "/additions", app.refillItem)
	router.HandlerFunc(http.MethodGet, "/additions/:id", app.listRefills)
	router.HandlerFunc(http.MethodPost, "/tags", app.insertTag)
	router.HandlerFunc(http.MethodGet, "/tags", app.getAllTags)
	router.HandlerFunc(http.MethodPost, "/tags/item", app.addItemTag)
	// list item tags
	router.HandlerFunc(http.MethodGet, "/tags/item/:id", app.listItemTags)

	router.HandlerFunc(http.MethodPost, "/user", app.registerUser)
	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.createAuthenticationToken)

	return app.recoverPanic(app.authenticate(router))
}
