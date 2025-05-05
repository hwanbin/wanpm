package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hwanbin/wanpm/internal/docs"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	router.Use(app.rateLimit)
	router.Use(app.enableCORS)
	router.Use(app.recoverPanic)

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	staticFS, err := docs.Assets()
	if err != nil {
		panic(err)
	}
	fs := http.StripPrefix("/docs/swagger-ui", http.FileServer(http.FS(staticFS)))
	router.Get("/docs/swagger-ui/*", fs.ServeHTTP)

	router.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write(docs.OpenAPISpec)
	})

	router.Get("/v1/healthcheck", app.healthcheckHandler)

	router.Post("/v1/user", app.registerUserHandler)
	router.Put("/v1/user/activate", app.activateUserHandler)
	router.Post("/v1/user/authenticate", app.authenticateUserHandler)

	if app.config.env == "remote-dev" {
		router.Use(app.authRequired)
	}

	router.Get("/v1/geocode/forward", app.forwardGeocodeHandler)

	router.Get("/v1/project", app.listProjectHandler)
	router.Post("/v1/project", app.createProjectHandler)
	router.Get("/v1/project/{id}", app.showProjectHandler)
	router.Patch("/v1/project/{id}", app.updateProjectHandler)
	router.Delete("/v1/project/{id}", app.deleteProjectHandler)

	router.Get("/v1/client", app.listClientHandler)
	router.Post("/v1/client", app.createClientHandler)
	router.Get("/v1/client/{id}", app.showClientHandler)
	router.Patch("/v1/client/{id}", app.updateClientHandler)
	router.Delete("/v1/client/{id}", app.deleteClientHandler)

	router.Post("/v1/proposal", app.createProposalHandler)
	router.Get("/v1/proposal/{id}", app.showProposalHandler)
	router.Patch("/v1/proposal/{id}", app.updateProposalHandler)
	router.Delete("/v1/proposal/{id}", app.deleteProposalHandler)

	router.Get("/v1/presigned-put", app.createPresignedPutUrlHandler)
	router.Get("/v1/presigned-get", app.createPresignedGetUrlHandler)
	router.Get("/v1/presigned-delete", app.createPresignedDeleteUrlHandler)

	router.Get("/v1/list-files", app.listFilesWithPrefixHandler)

	router.Post("/v1/activity", app.createActivityHandler)
	router.Get("/v1/activity/{id}", app.showActivityHandler)
	router.Get("/v1/activity", app.listActivityHandler)
	router.Patch("/v1/activity/{id}", app.updateActivityHandler)
	router.Delete("/v1/activity/{id}", app.deleteActivityHandler)

	router.Post("/v1/role", app.createRoleHandler)
	router.Get("/v1/role/{id}", app.showRoleHandler)
	router.Get("/v1/role", app.listRoleHandler)
	router.Patch("/v1/role/{id}", app.updateRoleHandler)
	router.Delete("/v1/role/{id}", app.deleteRoleHandler)

	router.Get("/v1/user/refresh", app.refreshTokenHandler)
	router.Get("/v1/user/logout", app.logoutHandler)
	router.Get("/v1/user/{id}", app.getUserHandler)
	router.Get("/v1/user", app.listUsersHandler)

	router.Post("/v1/timesheet", app.createTimesheetHandler)
	router.Get("/v1/timesheet/{id}", app.showTimesheetHandler)
	router.Get("/v1/timesheet", app.listTimesheetHandler)
	router.Patch("/v1/timesheet/{id}", app.updateTimesheetHandler)
	router.Delete("/v1/timesheet/{id}", app.deleteTimesheetHandler)
	
	return router
}
