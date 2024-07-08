package pgrest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// StartServer starts the PGRest server with the given configuration.
// It initializes the necessary resources, sets up the main handler,
// and listens for incoming HTTP requests on the specified port.
func StartServer(config Config) {
	defer CloseDBPools()

	log.Info(fmt.Sprintf("PGRest started, running on port %v", config.PGRest.Port))
	http.Handle("/", mainHandler(QueryHandler, config.Connections))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.PGRest.Port), nil))
}

// mainHandler is a HTTP handler function that wraps the provided handler function
// and performs various checks and validations before executing the handler.
// It checks the request method, validates the authorization header, decodes the
// incoming JSON payload, finds the requested database connection, checks if the
// user has access to the database, checks if the request is allowed from the origin,
// and finally executes the provided handler function.
func mainHandler(handler func(http.ResponseWriter, *http.Request, ConnectionConfig, RequestBody) error, connections []ConnectionConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Reset the request body with the original data
		r.Body = io.NopCloser(bytes.NewBuffer(body))

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
			if c.Name == requestBody.Connection {
				connection = c
				break
			}
		}

		// if connection auth is public, handle the request without authentication
		if connection.Auth == "public" {
			err = handler(w, r, connection, requestBody)
			if err != nil {
				log.Errorf("Error handling request: %v", err)
				handleError(w, err)
			}

			return
		}

		// Get the Authorization header
		clientID, token, err := getAuthHeader(r)
		if err != nil {
			handleError(w, err)
			return
		}

		// Check if the user has access to the requested connection
		user := getConnectionUser(clientID, connection)
		if user == nil {
			apiError := NewAPIError(http.StatusUnauthorized, "Unauthorized access for connection", nil)
			handleError(w, apiError)
			return
		}

		// Validate the auth token
		generatedToken := getHMACToken(string(body), user.ClientSecret)
		if generatedToken != token {
			apiError := NewAPIError(http.StatusUnauthorized, "Invalid token", nil)
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
			apiError := NewAPIError(http.StatusBadRequest, fmt.Sprintf("Requested connection '%s' not found", requestBody.Connection), nil)
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

// getAuthHeader extracts the clientID and HMAC from the Authorization header of an HTTP request.
// It expects the Authorization header to be in the format "Bearer base64(clientID:HMAC)".
// If the header is missing, invalid, or cannot be decoded, it returns an error.
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

	credentials := strings.Split(string(decodedBytes), ".")
	if len(credentials) != 2 {
		return "", "", NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	return credentials[0], credentials[1], nil
}

// getConnectionUser returns the UserConfig associated with the given client ID and API key from the provided ConnectionConfig.
// If no matching user is found, it returns nil.
func getConnectionUser(clientID string, connection ConnectionConfig) *UserConfig {
	for _, user := range connection.Users {
		if user.ClientID == clientID {
			return &user
		}
	}

	return nil
}

// getHMACToken generates an HMAC token for the given message using the provided secret.
// It uses the SHA256 hashing algorithm and encodes the resulting hash in base64 format.
func getHMACToken(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// handleError handles the error and sends an appropriate response to the client.
// If the error is of type *APIError, it sets the response status code and encodes the error as JSON.
// For unexpected errors, it sets the response status code to 500 and creates a new APIError with a generic message.
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
