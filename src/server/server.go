package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chim "github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/sogelink-research/pgrest/api/handlers"
	"github.com/sogelink-research/pgrest/api/middleware"
	"github.com/sogelink-research/pgrest/database"
	"github.com/sogelink-research/pgrest/settings"
)

// Start starts the PGRest server with the given configuration.
// It initializes the necessary resources, sets up the main handler,
// and listens for incoming HTTP requests on the specified port.
func Start(config settings.Config) {
	router := createRouter(config)
	server := &http.Server{Addr: fmt.Sprintf(":%v", config.PGRest.Port), Handler: router}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		log.Info("Stop signal received, shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 5*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Server stopped successfully")
		serverStopCtx()
	}()

	log.Info(fmt.Sprintf("PGRest started, running on port %v", config.PGRest.Port))
	defer database.CloseDBPools()

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

// createRouter creates and configures the router for the server.
// It sets up the necessary middleware and routes for handling API requests.
// The router is configured with the provided `config` settings.
func createRouter(config settings.Config) http.Handler {
	router := chi.NewRouter()
	router.Use(chim.Recoverer)
	router.Use(chim.Throttle(config.PGRest.MaxConcurrentRequests))

	router.NotFound(handlers.NotFoundHandler)
	router.Route("/api/{connection}/query", func(r chi.Router) {
		r.Use(middleware.CORSMiddleware(config.PGRest.CORS))
		r.Use(middleware.AuthMiddleware(config))
		r.Post("/", handlers.QueryHandler(config))
	})

	return router
}
