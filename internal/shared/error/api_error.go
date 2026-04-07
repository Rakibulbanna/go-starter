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

// NewAPIError creates a new API error
func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Status:     code,
		Message:    message,
	}
}

// NewAPIErrorWithDetails creates a new API error with additional details
func NewAPIErrorWithDetails(statusCode int, code, message string, details map[string]interface{}) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Status:     code,
		Message:    message,
		Details:    details,
	}
}

// NewAPIErrorWithError creates a new API error with an underlying error
func NewAPIErrorWithError(statusCode int, code, message string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Status:     code,
		Message:    message,
		Err:        err,
	}
}

// Common API Error Constructors

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) *APIError {
	return NewAPIError(http.StatusBadRequest, "BAD_REQUEST", message)
}

// BadRequestWithDetails creates a 400 Bad Request error with details
func BadRequestWithDetails(message string, details map[string]interface{}) *APIError {
	return NewAPIErrorWithDetails(http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *APIError {
	return NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *APIError {
	return NewAPIError(http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound creates a 404 Not Found error
func NotFound(message string) *APIError {
	return NewAPIError(http.StatusNotFound, "NOT_FOUND", message)
}

// NotFoundWithDetails creates a 404 Not Found error with details
func NotFoundWithDetails(message string, details map[string]interface{}) *APIError {
	return NewAPIErrorWithDetails(http.StatusNotFound, "NOT_FOUND", message, details)
}

// Conflict creates a 409 Conflict error
func Conflict(message string) *APIError {
	return NewAPIError(http.StatusConflict, "CONFLICT", message)
}

// UnprocessableEntity creates a 422 Unprocessable Entity error
func UnprocessableEntity(message string, details map[string]interface{}) *APIError {
	return NewAPIErrorWithDetails(http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message, details)
}

// InternalServerError creates a 500 Internal Server Error
func InternalServerError(message string) *APIError {
	return NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// InternalServerErrorWithError creates a 500 Internal Server Error with underlying error
func InternalServerErrorWithError(message string, err error) *APIError {
	return NewAPIErrorWithError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, err)
}

// ServiceUnavailable creates a 503 Service Unavailable error
func ServiceUnavailable(message string) *APIError {
	return NewAPIError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// ValidationError creates a validation error with field-specific details
func ValidationError(field, message string) *APIError {
	details := map[string]interface{}{
		"field": field,
		"error": message,
	}
	return NewAPIErrorWithDetails(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", details)
}

// MultipleValidationErrors creates a validation error with multiple field errors
func MultipleValidationErrors(errors map[string]string) *APIError {
	details := map[string]interface{}{
		"validation_errors": errors,
	}
	return NewAPIErrorWithDetails(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Multiple validation errors", details)
}

// ResourceNotFound creates a resource not found error
func ResourceNotFound(resourceType, identifier string) *APIError {
	message := fmt.Sprintf("%s with identifier '%s' not found", resourceType, identifier)
	details := map[string]interface{}{
		"resource_type": resourceType,
		"identifier":    identifier,
	}
	return NewAPIErrorWithDetails(http.StatusNotFound, "RESOURCE_NOT_FOUND", message, details)
}

// DuplicateResource creates a duplicate resource error
func DuplicateResource(resourceType, field, value string) *APIError {
	message := fmt.Sprintf("%s with %s '%s' already exists", resourceType, field, value)
	details := map[string]interface{}{
		"resource_type": resourceType,
		"field":         field,
		"value":         value,
	}
	return NewAPIErrorWithDetails(http.StatusConflict, "DUPLICATE_RESOURCE", message, details)
}

// InsufficientPermissions creates a permission error
func InsufficientPermissions(action, resource string) *APIError {
	message := fmt.Sprintf("Insufficient permissions to %s %s", action, resource)
	details := map[string]interface{}{
		"action":   action,
		"resource": resource,
	}
	return NewAPIErrorWithDetails(http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", message, details)
}

// RateLimitExceeded creates a rate limit error
func RateLimitExceeded(message string) *APIError {
	details := map[string]interface{}{
		"message": message,
	}
	return NewAPIErrorWithDetails(http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message, details)
}

// DatabaseError creates a database error
func DatabaseError(operation string, err error) *APIError {
	message := fmt.Sprintf("Database %s failed", operation)
	details := map[string]interface{}{
		"operation": operation,
	}
	return NewAPIErrorWithDetails(http.StatusInternalServerError, "DATABASE_ERROR", message, details)
}

// ExternalServiceError creates an external service error
func ExternalServiceError(service, operation string, err error) *APIError {
	message := fmt.Sprintf("%s service %s failed", service, operation)
	details := map[string]interface{}{
		"service":   service,
		"operation": operation,
	}
	return NewAPIErrorWithDetails(http.StatusServiceUnavailable, "EXTERNAL_SERVICE_ERROR", message, details)
}

func UnsupportedMediaType(message string) *APIError {
	return NewAPIError(http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", message)
}
