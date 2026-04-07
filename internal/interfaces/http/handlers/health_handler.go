package handlers

import (
	"net/http"
	"time"

	"skoolz/config"
	"skoolz/internal/shared/response"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	serviceName string
	version     string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	cfg := config.GetConfig()
	return &HealthHandler{
		serviceName: cfg.ServiceName,
		version:     cfg.Version,
	}
}

// HealthCheck handles the health check endpoint
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"status":    "ok",
		"service":   h.serviceName,
		"timestamp": time.Now(),
		"version":   h.version,
	}

	response.WriteOK(w, "Service health check completed", data)
}
