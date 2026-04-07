package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// status string used in the JSON envelope
const (
	statusSuccess = "success"
	statusError   = "error"
)

// baseResponse is the common envelope for every JSON response.
type baseResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// successResponse is the standard success payload.
type successResponse struct {
	baseResponse
	Data interface{} `json:"data,omitempty"`
}

// errorResponse is the standard error payload.
type errorResponse struct {
	baseResponse
	Error string `json:"error"`
}

// paginatedResponse is returned by list endpoints.
type paginatedResponse struct {
	baseResponse
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination metadata returned with list responses.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// writeJSON serialises any payload as JSON with the given status code.
func writeJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

// firstString picks the first string from opts, returning fallback if none.
func firstString(opts []any, fallback string) (string, []any) {
	if len(opts) > 0 {
		if s, ok := opts[0].(string); ok && s != "" {
			return s, opts[1:]
		}
	}
	return fallback, opts
}

// writeSuccess emits a success envelope. Variadic args may be:
//
//	(message string, data any)  // most common
//	(data any)                   // message defaults to "Success"
//	()                           // empty body
func writeSuccess(w http.ResponseWriter, statusCode int, defaultMessage string, opts ...any) {
	message := defaultMessage
	var data any

	if len(opts) > 0 {
		if msg, ok := opts[0].(string); ok && msg != "" {
			message = msg
			if len(opts) > 1 {
				data = opts[1]
			}
		} else {
			data = opts[0]
		}
	}

	writeJSON(w, statusCode, successResponse{
		baseResponse: baseResponse{
			Status:    statusSuccess,
			Message:   message,
			Timestamp: time.Now(),
		},
		Data: data,
	})
}

// writeError emits an error envelope.
func writeError(w http.ResponseWriter, statusCode int, defaultMessage, defaultCode string, opts ...any) {
	message, rest := firstString(opts, defaultMessage)
	code, _ := firstString(rest, defaultCode)

	writeJSON(w, statusCode, errorResponse{
		baseResponse: baseResponse{
			Status:    statusError,
			Message:   message,
			Timestamp: time.Now(),
		},
		Error: code,
	})
}

// WriteOK writes a 200 OK success response.
func WriteOK(w http.ResponseWriter, opts ...any) {
	writeSuccess(w, http.StatusOK, "Success", opts...)
}

// WriteCreated writes a 201 Created success response.
func WriteCreated(w http.ResponseWriter, opts ...any) {
	writeSuccess(w, http.StatusCreated, "Created", opts...)
}

// WriteBadRequest writes a 400 Bad Request error response.
func WriteBadRequest(w http.ResponseWriter, opts ...any) {
	writeError(w, http.StatusBadRequest, "Bad Request", "BAD_REQUEST", opts...)
}

// WriteNotFound writes a 404 Not Found error response.
func WriteNotFound(w http.ResponseWriter, opts ...any) {
	writeError(w, http.StatusNotFound, "Not Found", "NOT_FOUND", opts...)
}

// WriteInternalServerError writes a 500 Internal Server Error response.
func WriteInternalServerError(w http.ResponseWriter, opts ...any) {
	writeError(w, http.StatusInternalServerError, "Internal Server Error", "INTERNAL_SERVER_ERROR", opts...)
}

// WritePaginated writes a list response with pagination metadata.
//
// Signature: WritePaginated(w, statusCode, message, data, page, limit, total)
func WritePaginated(w http.ResponseWriter, statusCode int, message string, data interface{}, page, limit int, total int64) {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	writeJSON(w, statusCode, paginatedResponse{
		baseResponse: baseResponse{
			Status:    statusSuccess,
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
	})
}
