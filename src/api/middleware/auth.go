package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/sogelink-research/pgrest/api/handlers"
	"github.com/sogelink-research/pgrest/errors"
	"github.com/sogelink-research/pgrest/settings"
	"github.com/sogelink-research/pgrest/utils"
)

// AuthMiddleware is a middleware function that handles authentication for API requests.
// It takes a `config` parameter of type `settings.Config` which contains the configuration settings.
// The function returns a `func(http.Handler) http.Handler` which can be used as middleware in the API router.
// The middleware validates the authentication token, checks if the requested connection is accessible by the user,
// and performs additional origin checks for security.
// If the authentication is successful, the middleware calls the next handler in the chain.
// If any error occurs during the authentication process, it returns an appropriate error response.
func AuthMiddleware(config settings.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connectionName, err := utils.GetConnectionNameFromRequest(r)
			if err != nil {
				handlers.HandleError(w, err)
				return
			}

			// Find the requested database connection
			connection, err := config.GetConnectionConfig(connectionName)
			if err != nil {
				apiError := errors.NewAPIError(http.StatusBadRequest, fmt.Sprintf("Requested connection '%s' not found", connectionName), nil)
				handlers.HandleError(w, apiError)
				return
			}

			// if connection auth is public, handle the request without authentication
			if connection.Auth == "public" {
				next.ServeHTTP(w, r)
				return
			}

			// Get the request body data
			bodyString, err := utils.GetBodyString(r)
			if err != nil {
				handlers.HandleError(w, err)
				return
			}

			// Get the Authorization header
			clientID, token, err := getAuthHeader(r)
			if err != nil {
				handlers.HandleError(w, err)
				return
			}

			// Find the user from config
			user, ok := config.UsersLookup[clientID]
			if !ok {
				apiError := errors.NewAPIError(http.StatusUnauthorized, "User not found", nil)
				handlers.HandleError(w, apiError)
				return
			}

			// Validate the auth token
			generatedToken := getHMACToken(bodyString, user.ClientSecret)
			if generatedToken != token {
				apiError := errors.NewAPIError(http.StatusUnauthorized, "Invalid token", nil)
				handlers.HandleError(w, apiError)
				return
			}

			// if connection.Name is not in the user's connections
			if !utils.Contains(user.Connections, connection.Name) {
				apiError := errors.NewAPIError(http.StatusUnauthorized, "User has not access to requested connection", nil)
				handlers.HandleError(w, apiError)
				return
			}

			// Additional origin check when send from backend
			// Can easily be bypassed by setting the Origin header to a value
			// but can add a little bit of security
			if !config.PGRest.CORS.IsOriginAllowed(r.Header.Get("Origin")) {
				apiError := errors.NewAPIError(http.StatusUnauthorized, "Unauthorized access from origin", nil)
				handlers.HandleError(w, apiError)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// getAuthHeader extracts the clientID and HMAC from the Authorization header of an HTTP request.
// It expects the Authorization header to be in the format "Bearer base64(clientID:HMAC)".
// If the header is missing, invalid, or cannot be decoded, it returns an error.
func getAuthHeader(r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", errors.NewAPIError(http.StatusUnauthorized, "Missing Authorization header", nil)
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", "", errors.NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(headerParts[1])
	if err != nil {
		return "", "", errors.NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	credentials := strings.Split(string(decodedBytes), ".")
	if len(credentials) != 2 {
		return "", "", errors.NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", nil)
	}

	return credentials[0], credentials[1], nil
}

// getHMACToken generates an HMAC token for the given message using the provided secret.
// It uses the SHA256 hashing algorithm and encodes the resulting hash in base64 format.
func getHMACToken(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
