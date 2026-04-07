package handlers

import (
	"net/http"
	"skoolz/config"
	"skoolz/internal/shared/response"
	"time"
)

// WelcomeResponse represents the welcome response
type WelcomeResponse struct {
	Message        string    `json:"message"`
	Service        string    `json:"service"`
	Timestamp      time.Time `json:"timestamp"`
	ProjectVersion string    `json:"project_version"`
}

// WelcomeHandler handles welcome requests
type WelcomeHandler struct {
	serviceName string
}

// NewWelcomeHandler creates a new welcome handler
func NewWelcomeHandler() *WelcomeHandler {
	return &WelcomeHandler{
		serviceName: config.GetConfig().ServiceName,
	}
}

// Welcome handles the welcome endpoint
func (h *WelcomeHandler) Welcome(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"message":         "Welcome to the " + h.serviceName + " API!",
		"service":         h.serviceName,
		"project_version": config.GetConfig().Version,
	}

	response.WriteOK(w, data)
}
