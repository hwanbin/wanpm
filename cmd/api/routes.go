package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hwanbin/wanpm-api/internal/docs"
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

	router.Route("/v0", func(mux chi.Router) {
		// router.Use(app.authRequired)

		mux.Get("/healthcheck", app.healthcheckHandler)
		mux.Get("/geocode/forward", app.forwardGeocodeHandler)
		mux.Get("/project", app.listProjectHandler)
		mux.Post("/project", app.createProjectHandler)
		mux.Get("/project/{id}", app.showProjectHandler)
		mux.Patch("/project/{id}", app.updateProjectHandler)
		mux.Delete("/project/{id}", app.deleteProjectHandler)
		mux.Get("/client", app.listClientHandler)
		mux.Post("/client", app.createClientHandler)
		mux.Get("/client/{id}", app.showClientHandler)
		mux.Patch("/client/{id}", app.updateClientHandler)
		mux.Delete("/client/{id}", app.deleteClientHandler)
		mux.Post("/proposal", app.createProposalHandler)
		mux.Get("/proposal/{id}", app.showProposalHandler)
		mux.Patch("/proposal/{id}", app.updateProposalHandler)
		mux.Delete("/proposal/{id}", app.deleteProposalHandler)
		mux.Get("/presigned-put", app.createPresignedPutUrlHandler)
		mux.Get("/presigned-get", app.createPresignedGetUrlHandler)
		mux.Get("/presigned-delete", app.createPresignedDeleteUrlHandler)
		mux.Get("/list-files", app.listFilesWithPrefixHandler)

		mux.Post("/activity", app.createActivityHandler)
		mux.Get("/activity/{id}", app.showActivityHandler)
		mux.Get("/activity", app.listActivityHandler)
		mux.Patch("/activity/{id}", app.updateActivityHandler)
		mux.Delete("/activity/{id}", app.deleteActivityHandler)

		mux.Post("/role", app.createRoleHandler)
		mux.Get("/role/{id}", app.showRoleHandler)
		mux.Get("/role", app.listRoleHandler)
		mux.Patch("/role/{id}", app.updateRoleHandler)
		mux.Delete("/role/{id}", app.deleteRoleHandler)

		mux.Post("/timesheet", app.createTimesheetHandler)
		mux.Get("/timesheet/{id}", app.showTimesheetHandler)
		mux.Get("/timesheet", app.listTimesheetHandler)
		mux.Patch("/timesheet/{id}", app.updateTimesheetHandler)
		mux.Delete("/timesheet/{id}", app.deleteTimesheetHandler)

		mux.Post("/user", app.registerUserHandler)
		mux.Put("/user/activate", app.activateUserHandler)
		mux.Post("/user/authenticate", app.authenticateUserHandler)
		mux.Get("/user/refresh", app.refreshTokenHandler)
		mux.Get("/user/logout", app.logoutHandler)

		mux.Get("/user/{id}", app.getUserHandler)
		mux.Get("/user", app.listUsersHandler)
	})

	router.Route("/v1", func(mux chi.Router) {
		mux.Use(app.authRequired)
		mux.Get("/healthcheck", app.healthcheckHandler)
		mux.Get("/geocode/forward", app.forwardGeocodeHandler)
		mux.Get("/project", app.listProjectHandler)
		mux.Post("/project", app.createProjectHandler)
		mux.Get("/project/{id}", app.showProjectHandler)
		mux.Patch("/project/{id}", app.updateProjectHandler)
		mux.Delete("/project/{id}", app.deleteProjectHandler)
		mux.Get("/client", app.listClientHandler)
		mux.Post("/client", app.createClientHandler)
		mux.Get("/client/{id}", app.showClientHandler)
		mux.Patch("/client/{id}", app.updateClientHandler)
		mux.Delete("/client/{id}", app.deleteClientHandler)
		mux.Post("/proposal", app.createProposalHandler)
		mux.Get("/proposal/{id}", app.showProposalHandler)
		mux.Patch("/proposal/{id}", app.updateProposalHandler)
		mux.Delete("/proposal/{id}", app.deleteProposalHandler)
		mux.Get("/presigned-put", app.createPresignedPutUrlHandler)
		mux.Get("/presigned-get", app.createPresignedGetUrlHandler)
		mux.Get("/presigned-delete", app.createPresignedDeleteUrlHandler)
		mux.Get("/list-files", app.listFilesWithPrefixHandler)

		mux.Post("/activity", app.createActivityHandler)
		mux.Get("/activity/{id}", app.showActivityHandler)
		mux.Get("/activity", app.listActivityHandler)
		mux.Patch("/activity/{id}", app.updateActivityHandler)
		mux.Delete("/activity/{id}", app.deleteActivityHandler)

		mux.Post("/role", app.createRoleHandler)
		mux.Get("/role/{id}", app.showRoleHandler)
		mux.Get("/role", app.listRoleHandler)
		mux.Patch("/role/{id}", app.updateRoleHandler)
		mux.Delete("/role/{id}", app.deleteRoleHandler)

		mux.Post("/timesheet", app.createTimesheetHandler)
		mux.Get("/timesheet/{id}", app.showTimesheetHandler)
		mux.Get("/timesheet", app.listTimesheetHandler)
		mux.Patch("/timesheet/{id}", app.updateTimesheetHandler)
		mux.Delete("/timesheet/{id}", app.deleteTimesheetHandler)
	})
	// router.Get("/v1/healthcheck", app.healthcheckHandler)

	// router.Get("/v1/geocode/forward", app.forwardGeocodeHandler)

	// router.Get("/v1/project", app.listProjectHandler)
	// router.Post("/v1/project", app.createProjectHandler)
	// router.Get("/v1/project/{id}", app.showProjectHandler)
	// router.Patch("/v1/project/{id}", app.updateProjectHandler)
	// router.Delete("/v1/project/{id}", app.deleteProjectHandler)

	// router.Get("/v1/client", app.listClientHandler)
	// router.Post("/v1/client", app.createClientHandler)
	// router.Get("/v1/client/{id}", app.showClientHandler)
	// router.Patch("/v1/client/{id}", app.updateClientHandler)
	// router.Delete("/v1/client/{id}", app.deleteClientHandler)

	// router.Post("/v1/proposal", app.createProposalHandler)
	// router.Get("/v1/proposal/{id}", app.showProposalHandler)
	// router.Patch("/v1/proposal/{id}", app.updateProposalHandler)
	// router.Delete("/v1/proposal/{id}", app.deleteProposalHandler)

	// router.Get("/v1/presigned-put", app.createPresignedPutUrlHandler)
	// router.Get("/v1/presigned-get", app.createPresignedGetUrlHandler)
	// router.Get("/v1/presigned-delete", app.createPresignedDeleteUrlHandler)

	// router.Get("/v1/list-files", app.listFilesWithPrefixHandler)

	router.Post("/v1/user", app.registerUserHandler)
	router.Put("/v1/user/activate", app.activateUserHandler)
	router.Post("/v1/user/authenticate", app.authenticateUserHandler)
	router.Get("/v1/user/refresh", app.refreshTokenHandler)
	router.Get("/v1/user/logout", app.logoutHandler)

	router.Get("/v1/user/{id}", app.getUserHandler)
	router.Get("/v1/user", app.listUsersHandler)

	return router
}
