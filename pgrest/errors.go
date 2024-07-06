package pgrest

import "net/http"

// APIError defines a custom error type with a status code, message, and details
type APIError struct {
	StatusCode int     `json:"status"`
	StatusText string  `json:"statusText"`
	Message    string  `json:"error"`
	Details    *string `json:"details,omitempty"`
}

// Implement the Error method to satisfy the error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new APIError with the given status code, message, and details
func NewAPIError(statusCode int, message string, details *string) *APIError {
	if message == "" {
		message = http.StatusText(statusCode)
	}

	return &APIError{
		StatusCode: statusCode,
		StatusText: http.StatusText(statusCode),
		Message:    message,
		Details:    details,
	}
}
