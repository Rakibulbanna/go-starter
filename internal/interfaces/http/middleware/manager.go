package middleware

import (
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

type Manager struct {
	globalMiddlewares []Middleware
}

// NewManager creates a new middleware manager with common middleware
func NewManager() *Manager {
	manager := &Manager{
		globalMiddlewares: make([]Middleware, 0),
	}

	// Add common middleware by default
	manager.UseDefault()

	return manager
}

// UseDefault adds common middleware that should be applied to all requests
func (m *Manager) UseDefault() *Manager {
	return m.Use(
		RecoveryMiddleware,                        // Panic recovery
		RequestIDMiddleware,                       // Request ID generation
		LoggingMiddleware,                         // Request logging
		SecurityHeadersMiddleware,                 // Security headers
		CORSMiddleware,                            // CORS handling
		TimeoutMiddleware(30*time.Second),         // Request timeout
		RateLimitMiddleware(60),                   // Rate limiting (60 requests per minute)
		ContentTypeMiddleware("application/json"), // Content type validation
	)
}

// Use adds middleware to the global middleware stack
func (m *Manager) Use(middlewares ...Middleware) *Manager {
	m.globalMiddlewares = append(m.globalMiddlewares, middlewares...)
	return m
}

// With applies middleware to a specific handler
func (m *Manager) With(handler http.Handler, middlewares ...Middleware) http.Handler {
	var h http.Handler
	h = handler

	// Apply route-specific middleware first
	for _, m := range middlewares {
		h = m(h)
	}

	// Apply global middleware last
	for _, m := range m.globalMiddlewares {
		h = m(h)
	}

	return h
}
