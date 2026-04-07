package handlers

import (
	"net/http"
	"time"

	"skoolz/config"
	"skoolz/internal/infrastructure/messaging"
	"skoolz/internal/shared/response"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]interface{} `json:"services,omitempty"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	serviceName      string
	version          string
	messagingFactory *messaging.MessagingFactory
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	serviceName := config.GetConfig().ServiceName
	version := config.GetConfig().Version
	return &HealthHandler{
		serviceName: serviceName,
		version:     version,
	}
}

// SetMessagingFactory sets the messaging factory for health checks
func (h *HealthHandler) SetMessagingFactory(factory *messaging.MessagingFactory) {
	h.messagingFactory = factory
}

// HealthCheck handles the health check endpoint
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]interface{})
	overallStatus := "ok"

	// Check messaging services if factory is available
	if h.messagingFactory != nil {
		if err := h.messagingFactory.HealthCheck(); err != nil {
			services["messaging"] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
			overallStatus = "degraded"
		} else {
			services["messaging"] = map[string]interface{}{
				"status": "ok",
			}
		}
	}

	data := map[string]interface{}{
		"status":    overallStatus,
		"service":   h.serviceName,
		"timestamp": time.Now(),
		"version":   h.version,
		"services":  services,
	}

	response.WriteOK(w, "Service health check completed", data)
}
