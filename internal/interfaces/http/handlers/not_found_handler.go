package handlers

import (
	"net/http"
	"skoolz/internal/shared/response"
)

// NotFoundHandler handles 404 Not Found requests
type NotFoundHandler struct{}

// NewNotFoundHandler creates a new not found handler
func NewNotFoundHandler() *NotFoundHandler {
	return &NotFoundHandler{}
}

// NotFound handles 404 Not Found requests
func (h *NotFoundHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	response.WriteNotFound(w, "The requested resource was not found", "RESOURCE_NOT_FOUND")
}
