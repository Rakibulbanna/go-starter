package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// ResponseStatus represents the status of a response
type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "success"
	StatusError   ResponseStatus = "error"
	StatusWarning ResponseStatus = "warning"
	StatusInfo    ResponseStatus = "info"
)

// BaseResponse represents the base structure for all API responses
type BaseResponse struct {
	Status    ResponseStatus `json:"status"`
	Message   string         `json:"message"`
	Timestamp time.Time      `json:"timestamp"`
	RequestID string         `json:"request_id,omitempty"`
}

// SuccessResponse represents a successful response with data
type SuccessResponse struct {
	BaseResponse
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	BaseResponse
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	BaseResponse
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// ResponseBuilder provides methods to build different types of responses
type ResponseBuilder struct{}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{}
}

// Success creates a success response
func (rb *ResponseBuilder) Success(message string, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		BaseResponse: BaseResponse{
			Status:    StatusSuccess,
			Message:   message,
			Timestamp: time.Now(),
		},
		Data: data,
	}
}

// Error creates an error response
func (rb *ResponseBuilder) Error(message, errorCode string, details map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			Status:    StatusError,
			Message:   message,
			Timestamp: time.Now(),
		},
		Error:   errorCode,
		Details: details,
	}
}

// Paginated creates a paginated response
func (rb *ResponseBuilder) Paginated(message string, data interface{}, page, limit int, total int64) *PaginatedResponse {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginatedResponse{
		BaseResponse: BaseResponse{
			Status:    StatusSuccess,
			Message:   message,
			Timestamp: time.Now(),
		},
		Data: data,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}
}

// WithRequestID adds a request ID to the response
func (rb *ResponseBuilder) WithRequestID(response interface{}, requestID string) {
	switch r := response.(type) {
	case *SuccessResponse:
		r.RequestID = requestID
	case *ErrorResponse:
		r.RequestID = requestID
	case *PaginatedResponse:
		r.RequestID = requestID
	}
}

// Helper functions for common response patterns
func CreateSuccessResponse(message string, data interface{}) *SuccessResponse {
	return NewResponseBuilder().Success(message, data)
}

func CreateErrorResponse(message, errorCode string) *ErrorResponse {
	return NewResponseBuilder().Error(message, errorCode, nil)
}

func CreateErrorResponseWithDetails(message, errorCode string, details map[string]interface{}) *ErrorResponse {
	return NewResponseBuilder().Error(message, errorCode, details)
}

func CreatePaginatedResponse(message string, data interface{}, page, limit int, total int64) *PaginatedResponse {
	return NewResponseBuilder().Paginated(message, data, page, limit, total)
}

// WriteSuccess writes a success response to the HTTP writer
func WriteSuccess(w http.ResponseWriter, statusCode int, opts ...interface{}) {
	message := "Success"
	var data interface{} = nil

	if len(opts) > 0 {
		// If first parameter is a string, treat it as message
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
			// If there's a second parameter, it's data
			if len(opts) > 1 {
				data = opts[1]
			}
		} else {
			// If first parameter is not a string, treat it as data
			data = opts[0]
		}
	}

	response := CreateSuccessResponse(message, data)
	writeJSON(w, statusCode, response)
}

// WriteError writes an error response to the HTTP writer
func WriteError(w http.ResponseWriter, statusCode int, opts ...interface{}) {
	message := "Error"
	errorCode := "ERROR"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	response := CreateErrorResponse(message, errorCode)
	writeJSON(w, statusCode, response)
}

// WriteErrorWithDetails writes an error response with details to the HTTP writer
func WriteErrorWithDetails(w http.ResponseWriter, statusCode int, opts ...interface{}) {
	message := "Error"
	errorCode := "ERROR"
	var details map[string]interface{} = nil

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}
	if len(opts) > 2 {
		if det, ok := opts[2].(map[string]interface{}); ok {
			details = det
		}
	}

	response := CreateErrorResponseWithDetails(message, errorCode, details)
	writeJSON(w, statusCode, response)
}

// WritePaginated writes a paginated response to the HTTP writer
func WritePaginated(w http.ResponseWriter, statusCode int, opts ...interface{}) {
	message := "Success"
	var data interface{} = nil
	page := 1
	limit := 10
	var total int64 = 0

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		data = opts[1]
	}
	if len(opts) > 2 {
		if p, ok := opts[2].(int); ok {
			page = p
		}
	}
	if len(opts) > 3 {
		if l, ok := opts[3].(int); ok {
			limit = l
		}
	}
	if len(opts) > 4 {
		if t, ok := opts[4].(int64); ok {
			total = t
		}
	}

	response := CreatePaginatedResponse(message, data, page, limit, total)
	writeJSON(w, statusCode, response)
}

// WriteJSON writes any JSON response to the HTTP writer
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	writeJSON(w, statusCode, data)
}

// writeJSON is a helper function that writes JSON responses
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Common HTTP status responses
func WriteOK(w http.ResponseWriter, opts ...any) {
	message := "Success"
	var data any = nil

	if len(opts) > 0 {
		// If first parameter is a string, treat it as message
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
			// If there's a second parameter, it's data
			if len(opts) > 1 {
				data = opts[1]
			}
		} else {
			// If first parameter is not a string, treat it as data
			data = opts[0]
		}
	}

	WriteSuccess(w, http.StatusOK, message, data)
}

func WriteCreated(w http.ResponseWriter, opts ...any) {
	message := "Created"
	var data any = nil

	if len(opts) > 0 {
		// If first parameter is a string, treat it as message
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
			// If there's a second parameter, it's data
			if len(opts) > 1 {
				data = opts[1]
			}
		} else {
			// If first parameter is not a string, treat it as data
			data = opts[0]
		}
	}

	WriteSuccess(w, http.StatusCreated, message, data)
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteBadRequest(w http.ResponseWriter, opts ...any) {
	message := "Bad Request"
	errorCode := "BAD_REQUEST"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	WriteError(w, http.StatusBadRequest, message, errorCode)
}

func WriteUnauthorized(w http.ResponseWriter, opts ...any) {
	message := "Unauthorized"
	errorCode := "UNAUTHORIZED"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	WriteError(w, http.StatusUnauthorized, message, errorCode)
}

func WriteForbidden(w http.ResponseWriter, opts ...any) {
	message := "Forbidden"
	errorCode := "FORBIDDEN"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	WriteError(w, http.StatusForbidden, message, errorCode)
}

func WriteNotFound(w http.ResponseWriter, opts ...any) {
	message := "Not Found"
	errorCode := "NOT_FOUND"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	WriteError(w, http.StatusNotFound, message, errorCode)
}

func WriteConflict(w http.ResponseWriter, opts ...any) {
	message := "Conflict"
	errorCode := "CONFLICT"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	WriteError(w, http.StatusConflict, message, errorCode)
}

func WriteUnprocessableEntity(w http.ResponseWriter, opts ...any) {
	message := "Unprocessable Entity"
	errorCode := "UNPROCESSABLE_ENTITY"
	var details map[string]any = nil

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}
	if len(opts) > 2 {
		if det, ok := opts[2].(map[string]interface{}); ok {
			details = det
		}
	}

	WriteErrorWithDetails(w, http.StatusUnprocessableEntity, message, errorCode, details)
}

func WriteInternalServerError(w http.ResponseWriter, opts ...any) {
	message := "Internal Server Error"
	errorCode := "INTERNAL_SERVER_ERROR"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	WriteError(w, http.StatusInternalServerError, message, errorCode)
}

func WriteServiceUnavailable(w http.ResponseWriter, opts ...any) {
	message := "Service Unavailable"
	errorCode := "SERVICE_UNAVAILABLE"

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
		}
	}
	if len(opts) > 1 {
		if code, ok := opts[1].(string); ok && code != "" {
			errorCode = code
		}
	}

	WriteError(w, http.StatusServiceUnavailable, message, errorCode)
}
