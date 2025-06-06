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

	router.Get("/v1/healthcheck", app.healthcheckHandler)

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

	return router
}
