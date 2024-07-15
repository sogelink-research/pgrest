package utils

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sogelink-research/pgrest/errors"
)

// Contains checks if a given string is present in a slice of strings.
// It returns true if the string is found, otherwise false.
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// GetConnectionNameFromRequest retrieves the connection name from the given HTTP request.
// It expects the connection name to be present as a path variable named "connection".
// If the connection name is not found or empty, it returns an error.
// The error returned is of type APIError with a status code of http.StatusBadRequest and a message indicating the absence of the connection name in the request.
func GetConnectionNameFromRequest(r *http.Request) (string, error) {
	connection := chi.URLParam(r, "connection")

	if connection == "" {
		return "", errors.NewAPIError(http.StatusBadRequest, "Connection name not found in request", nil)
	}

	return connection, nil
}
