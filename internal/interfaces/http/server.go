package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"skoolz/internal/infrastructure/container"
	"skoolz/internal/interfaces/http/routes"
	"skoolz/internal/logger"
	"strconv"
	"syscall"
	"time"

	"go.elastic.co/apm/module/apmhttp"
)

// Server represents the HTTP server
type Server struct {
	server    *http.Server
	container *container.Container
	done      chan struct{}
}

// NewServer creates a new HTTP server
func NewServer() *Server {
	// Get the container instance
	container := container.GetContainer()

	// Create router
	mux := http.NewServeMux()
	handler := routes.SetupRoutes(mux)

	// Wrap handler with APM monitoring
	apmHandler := apmhttp.Wrap(handler)

	// Get config for server settings
	conf := container.GetConfig()

	return &Server{
		server: &http.Server{
			Addr:         ":" + strconv.Itoa(conf.HttpPort),
			Handler:      apmHandler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		container: container,
		done:      make(chan struct{}),
	}
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	// Create a channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests in a separate goroutine.
	go func() {
		startupLog := logger.NewStartupLogger(slog.Default())
		startupLog.Service("Server starting", "port", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErrors:
		slog.Error("Server error occurred", "error", err)
		return err
	case <-s.shutdown():
		slog.Info("Shutdown signal received, starting graceful shutdown")
		return s.gracefulShutdown()
	}
}

// gracefulShutdown performs a graceful shutdown of the server
func (s *Server) gracefulShutdown() error {
	s.container.Lock()
	defer s.container.Unlock()

	// Close the done channel to signal shutdown
	close(s.done)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("Initiating graceful shutdown", "timeout", "30s")

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed, forcing close", "error", err)

		// Force close if graceful shutdown fails
		if closeErr := s.server.Close(); closeErr != nil {
			return fmt.Errorf("failed to force close server: %w", closeErr)
		}
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	// Cleanup container connections
	if err := s.container.Close(); err != nil {
		slog.Error("Failed to close container connections", "error", err)
	} else {
		slog.Info("Container connections closed successfully")
	}

	slog.Info("Server shutdown completed successfully")
	return nil
}

// shutdown returns a channel that will be closed when shutdown should occur.
func (s *Server) shutdown() chan os.Signal {
	// Make channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)

	// Listen for multiple shutdown signals
	signal.Notify(shutdown,
		os.Interrupt,    // Ctrl+C
		syscall.SIGTERM, // Termination signal
		syscall.SIGINT,  // Interrupt signal
		syscall.SIGQUIT, // Quit signal
	)

	return shutdown
}
