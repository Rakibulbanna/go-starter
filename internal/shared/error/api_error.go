package AppError

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIError represents a structured API error
type APIError struct {
	StatusCode int                    `json:"-"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Err        error                  `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Err
}

// WriteToResponse writes the API error to an HTTP response writer
func (e *APIError) WriteToResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)

	response := map[string]interface{}{
		"status":    "error",
		"info":      e.Status,
		"message":   e.Message,
		"timestamp": time.Now(),
	}

	if e.Details != nil {
		response["details"] = e.Details
	}

	json.NewEncoder(w).Encode(response)
}

// newAPIError creates a new API error
func newAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Status:     code,
		Message:    message,
	}
}

// InternalServerError creates a 500 Internal Server Error
func InternalServerError(message string) *APIError {
	return newAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// RateLimitExceeded creates a 429 Too Many Requests error
func RateLimitExceeded(message string) *APIError {
	return newAPIError(http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message)
}

// UnsupportedMediaType creates a 415 Unsupported Media Type error
func UnsupportedMediaType(message string) *APIError {
	return newAPIError(http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", message)
}
