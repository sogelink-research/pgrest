package pgrest

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func StartServer(config Config) {
	http.Handle("/", mainHandler(QueryHandler, config.Connections))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.PGRest.Port), nil))
}

func mainHandler(handler func(http.ResponseWriter, *http.Request, ConnectionConfig, RequestBody) error, connections []ConnectionConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestBody RequestBody

		// Decode the incoming JSON payload
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			details := err.Error()
			apiError := NewAPIError(http.StatusBadRequest, "Invalid request body", &details)
			handleError(w, apiError)
			return
		}

		// Find the requested database connection
		var connection ConnectionConfig
		for _, c := range connections {
			if c.Name == requestBody.Database {
				connection = c
				break
			}
		}

		// if no connection is found, return an error
		if connection.Name == "" {
			apiError := NewAPIError(http.StatusBadRequest, fmt.Sprintf("Requested database '%s' not found", requestBody.Database), nil)
			handleError(w, apiError)
			return
		}

		err := handler(w, r, connection, requestBody)
		if err != nil {
			log.Errorf("Error handling request: %v", err)
			handleError(w, err)
		}
	})
}

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(apiErr.StatusCode)
			json.NewEncoder(w).Encode(apiErr)
			return
		}

		// Fallback for unexpected errors
		w.WriteHeader(http.StatusInternalServerError)
		response := NewAPIError(http.StatusInternalServerError, "An unexpected error occurred", nil)
		json.NewEncoder(w).Encode(response)
	}
}
