package pgrest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func StartServer(config Config) {
	defer CloseDBPools()

	http.Handle("/", mainHandler(QueryHandler, config.Connections))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.PGRest.Port), nil))
}

func mainHandler(handler func(http.ResponseWriter, *http.Request, ConnectionConfig, RequestBody) error, connections []ConnectionConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Get the Authorization header
		clientID, apiKey, err := getAuthHeader(r)
		if err != nil {
			handleError(w, err)
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

		// Check if the user has access to the requested database
		user := getConnectionUser(clientID, apiKey, connection)

		if user == nil {
			apiError := NewAPIError(http.StatusUnauthorized, "Unauthorized access to database", nil)
			handleError(w, apiError)
			return
		}

		// Check if the request is allowed from the origin
		if !user.CORS.isOriginAllowed(r.Header.Get("Origin")) {
			apiError := NewAPIError(http.StatusUnauthorized, "Unauthorized access from origin", nil)
			handleError(w, apiError)
			return
		}

		// if no connection is found, return an error
		if connection.Name == "" {
			apiError := NewAPIError(http.StatusBadRequest, fmt.Sprintf("Requested database '%s' not found", requestBody.Database), nil)
			handleError(w, apiError)
			return
		}

		err = handler(w, r, connection, requestBody)
		if err != nil {
			log.Errorf("Error handling request: %v", err)
			handleError(w, err)
		}
	})
}

// function to get Authorization header from request Bearar username:apikey
func getAuthHeader(r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", NewAPIError(http.StatusUnauthorized, "Missing Authorization header", nil)
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", "", NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	if headerParts[0] != "Bearer" {
		return "", "", NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(headerParts[1])
	if err != nil {
		return "", "", NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	credentials := strings.Split(string(decodedBytes), ":")
	if len(credentials) != 2 {
		return "", "", NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	return credentials[0], credentials[1], nil
}

// function to check if the user has access to the requested database
func getConnectionUser(clientID, apiKey string, connection ConnectionConfig) *UserConfig {
	for _, user := range connection.Users {
		if user.ClientID == clientID && user.APIKey == apiKey {
			return &user
		}
	}

	return nil
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
