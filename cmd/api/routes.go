package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/geocode/forward", app.forwardGeocodeHandler)

	router.HandlerFunc(http.MethodGet, "/v1/project", app.listProjectHandler)
	router.HandlerFunc(http.MethodPost, "/v1/project", app.createProjectHandler)
	router.HandlerFunc(http.MethodGet, "/v1/project/:id", app.showProjectHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/project/:id", app.updateProjectHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/project/:id", app.deleteProjectHandler)

	// router.HandlerFunc(http.MethodGet, "/v1/project", app.requireActivatedUser(app.listProjectHandler))
	// router.HandlerFunc(http.MethodPost, "/v1/project", app.requireActivatedUser(app.createProjectHandler))
	// router.HandlerFunc(http.MethodGet, "/v1/project/:id", app.requireActivatedUser(app.showProjectHandler))
	// router.HandlerFunc(http.MethodPatch, "/v1/project/:id", app.requireActivatedUser(app.updateProjectHandler))
	// router.HandlerFunc(http.MethodDelete, "/v1/project/:id", app.requireActivatedUser(app.deleteProjectHandler))

	// router.HandlerFunc(http.MethodGet, "/v1/project", app.requirePermission("project:read", app.listProjectHandler))
	// router.HandlerFunc(http.MethodPost, "/v1/project", app.requirePermission("project:write", app.createProjectHandler))
	// router.HandlerFunc(http.MethodGet, "/v1/project/:id", app.requirePermission("project:read", app.showProjectHandler))
	// router.HandlerFunc(http.MethodPatch, "/v1/project/:id", app.requirePermission("project:write", app.updateProjectHandler))
	// router.HandlerFunc(http.MethodDelete, "/v1/project/:id", app.requirePermission("project:write", app.deleteProjectHandler))

	router.HandlerFunc(http.MethodGet, "/v1/client", app.listClientHandler)
	router.HandlerFunc(http.MethodPost, "/v1/client", app.createClientHandler)
	router.HandlerFunc(http.MethodGet, "/v1/client/:id", app.showClientHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/client/:id", app.updateClientHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/client/:id", app.deleteClientHandler)

	router.HandlerFunc(http.MethodPost, "/v1/proposal", app.createProposalHandler)
	router.HandlerFunc(http.MethodGet, "/v1/proposal/:id", app.showProposalHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/proposal/:id", app.updateProposalHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/proposal/:id", app.deleteProposalHandler)

	router.HandlerFunc(http.MethodPost, "/v1/user", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/user/activate", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/token/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
	// return app.recoverPanic(app.rateLimit(app.authenticate(router)))
	// return app.recoverPanic(app.rateLimit(router))
}
