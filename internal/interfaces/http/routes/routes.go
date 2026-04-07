package routes

import (
	"net/http"
	"skoolz/internal/interfaces/http/handlers"
	"skoolz/internal/interfaces/http/middleware"
)

// Router handles HTTP routing
// SetupRoutes configures all routes with global middleware
func SetupRoutes(mux *http.ServeMux) http.Handler {

	// Initialize middleware manager
	manager := middleware.NewManager()

	// Initialize handlers
	welcomeHandler := handlers.NewWelcomeHandler()
	healthHandler := handlers.NewHealthHandler()
	notFoundHandler := handlers.NewNotFoundHandler()
	taskHandler := handlers.NewTaskHandler()

	// Define routes with middleware
	mux.Handle("GET /api", manager.With(http.HandlerFunc(welcomeHandler.Welcome)))
	mux.Handle("GET /api/v1", manager.With(http.HandlerFunc(welcomeHandler.Welcome)))
	mux.Handle("GET /health", manager.With(http.HandlerFunc(healthHandler.HealthCheck)))

	// Task CRUD routes
	mux.Handle("POST /api/v1/tasks", manager.With(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("GET /api/v1/tasks", manager.With(http.HandlerFunc(taskHandler.List)))
	mux.Handle("GET /api/v1/tasks/{id}", manager.With(http.HandlerFunc(taskHandler.Get)))
	mux.Handle("PUT /api/v1/tasks/{id}", manager.With(http.HandlerFunc(taskHandler.Update)))
	mux.Handle("DELETE /api/v1/tasks/{id}", manager.With(http.HandlerFunc(taskHandler.Delete)))

	// Catch-all route for 404 Not Found
	mux.HandleFunc("/", notFoundHandler.NotFound)

	return mux
}
