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

	router.Group(func(mux chi.Router) {
		if app.config.env == "remote-dev" {
			mux.Use(app.authRequired)
		}

		mux.Get("/v1/geocode/forward", app.forwardGeocodeHandler)

		mux.Get("/v1/project", app.listProjectHandler)
		mux.Post("/v1/project", app.createProjectHandler)
		mux.Get("/v1/project/{id}", app.showProjectHandler)
		mux.Patch("/v1/project/{id}", app.updateProjectHandler)
		mux.Delete("/v1/project/{id}", app.deleteProjectHandler)

		mux.Get("/v1/client", app.listClientHandler)
		mux.Post("/v1/client", app.createClientHandler)
		mux.Get("/v1/client/{id}", app.showClientHandler)
		mux.Patch("/v1/client/{id}", app.updateClientHandler)
		mux.Delete("/v1/client/{id}", app.deleteClientHandler)

		mux.Post("/v1/proposal", app.createProposalHandler)
		mux.Get("/v1/proposal/{id}", app.showProposalHandler)
		mux.Patch("/v1/proposal/{id}", app.updateProposalHandler)
		mux.Delete("/v1/proposal/{id}", app.deleteProposalHandler)

		mux.Get("/v1/presigned-put", app.createPresignedPutUrlHandler)
		mux.Get("/v1/presigned-get", app.createPresignedGetUrlHandler)
		mux.Get("/v1/presigned-delete", app.createPresignedDeleteUrlHandler)

		mux.Get("/v1/list-files", app.listFilesWithPrefixHandler)

		mux.Post("/v1/activity", app.createActivityHandler)
		mux.Get("/v1/activity/{id}", app.showActivityHandler)
		mux.Get("/v1/activity", app.listActivityHandler)
		mux.Patch("/v1/activity/{id}", app.updateActivityHandler)
		mux.Delete("/v1/activity/{id}", app.deleteActivityHandler)

		mux.Post("/v1/role", app.createRoleHandler)
		mux.Get("/v1/role/{id}", app.showRoleHandler)
		mux.Get("/v1/role", app.listRoleHandler)
		mux.Patch("/v1/role/{id}", app.updateRoleHandler)
		mux.Delete("/v1/role/{id}", app.deleteRoleHandler)

		mux.Get("/v1/user/refresh", app.refreshTokenHandler)
		mux.Get("/v1/user/logout", app.logoutHandler)
		mux.Get("/v1/user/{id}", app.getUserHandler)
		mux.Get("/v1/user", app.listUsersHandler)

		mux.Post("/v1/timesheet", app.createTimesheetHandler)
		mux.Get("/v1/timesheet/{id}", app.showTimesheetHandler)
		mux.Get("/v1/timesheet", app.listTimesheetHandler)
		mux.Patch("/v1/timesheet/{id}", app.updateTimesheetHandler)
		mux.Delete("/v1/timesheet/{id}", app.deleteTimesheetHandler)
	})

	return router
}
