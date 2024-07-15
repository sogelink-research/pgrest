package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/sogelink-research/pgrest/api/handlers"
	"github.com/sogelink-research/pgrest/api/middleware"
	"github.com/sogelink-research/pgrest/database"
	"github.com/sogelink-research/pgrest/errors"
	"github.com/sogelink-research/pgrest/settings"
)

// StartServer starts the PGRest server with the given configuration.
// It initializes the necessary resources, sets up the main handler,
// and listens for incoming HTTP requests on the specified port.
func StartServer(conf settings.Config) {
	router := chi.NewRouter()

	// Add routes
	router.Route("/api/{connection}/query", func(r chi.Router) {
		// Add middleware
		r.Use(middleware.CORSMiddleware(conf.PGRest.CORS))
		r.Use(middleware.AuthMiddleware(conf))
		r.Post("/", handlers.QueryHandler(conf))
	})

	// Handle not found routes
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		details := fmt.Sprintf("Path '%s' not found", r.URL.Path)
		error := errors.NewAPIError(http.StatusNotFound, "Not found", &details)
		handlers.HandleError(w, error)
	})

	defer database.CloseDBPools()

	// Start the server
	log.Info(fmt.Sprintf("PGRest started, running on port %v", conf.PGRest.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", conf.PGRest.Port), router))
}
