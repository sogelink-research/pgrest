package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

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
	router := mux.NewRouter()

	// Add routes
	router.HandleFunc("/api/{connection}/query", handlers.QueryHandler(conf)).Methods("POST")

	// Add middleware
	router.Use(middleware.CORSMiddleware(conf.PGRest.CORS))
	router.Use(middleware.AuthMiddleware(conf))

	// Handle not found routes
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		details := fmt.Sprintf("Path '%s' not found", r.URL.Path)
		error := errors.NewAPIError(http.StatusNotFound, "Not found", &details)
		handlers.HandleError(w, error)
	})

	defer database.CloseDBPools()

	// Start the server
	log.Info(fmt.Sprintf("PGRest started, running on port %v", conf.PGRest.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", conf.PGRest.Port), router))
}
