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
	router.HandlerFunc(http.MethodGet, "/items", app.requirePermission("read", app.getItems))
	router.HandlerFunc(http.MethodGet, "/items/:id", app.requirePermission("read", app.getItem))
	router.HandlerFunc(http.MethodPost, "/items", app.requirePermission("write", app.addItem))
	router.HandlerFunc(http.MethodPut, "/items/:id", app.requirePermission("write", app.updateItem))
	router.HandlerFunc(http.MethodDelete, "/items/:id", app.requirePermission("write", app.deleteItem))
	router.HandlerFunc(http.MethodGet, "/issues/:id", app.requirePermission("read", app.listIssues))
	router.HandlerFunc(http.MethodPost, "/issues", app.requirePermission("issue", app.addIssue))
	router.HandlerFunc(http.MethodPost, "/removals", app.requirePermission("issue", app.addRemoval))
	router.HandlerFunc(http.MethodGet, "/removals/:id", app.requirePermission("read", app.listRemovals))
	router.HandlerFunc(http.MethodPost, "/additions", app.requirePermission("write", app.refillItem))
	router.HandlerFunc(http.MethodGet, "/additions/:id", app.requirePermission("read", app.listRefills))
	router.HandlerFunc(http.MethodPost, "/tags", app.requirePermission("write", app.insertTag))
	router.HandlerFunc(http.MethodGet, "/tags", app.requirePermission("read", app.getAllTags))
	router.HandlerFunc(http.MethodPost, "/tags/item", app.requirePermission("write", app.addItemTag))
	router.HandlerFunc(http.MethodGet, "/tags/item/:id", app.requirePermission("read", app.listItemTags))
	router.HandlerFunc(http.MethodPost, "/users", app.registerUser)
	router.HandlerFunc(http.MethodGet, "/users", app.requireAdmin(app.getAllUsers))
	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.createAuthenticationToken)
	router.HandlerFunc(http.MethodPost, "/tokens/validate", app.validateToken)
	router.HandlerFunc(http.MethodPost, "/users/permissions", app.requireAdmin(app.updatePermission))
	router.HandlerFunc(http.MethodGet, "/users/permissions/:id", app.requireAdmin(app.getUserPermissionById))

	return app.recoverPanic(app.authenticate(router))
}
