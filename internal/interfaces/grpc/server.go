//go:build grpc

package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"skoolz/internal/infrastructure/container"
	"skoolz/internal/interfaces/grpc/middleware"
	"skoolz/internal/interfaces/grpc/routes"
	"strconv"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents the gRPC server
type Server struct {
	server    *grpc.Server
	container *container.Container
}

// NewServer creates a new gRPC server
func NewServer() *Server {
	// Get the container instance
	container := container.GetContainer()

	// Create gRPC server with options
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryInterceptor),
	)

	// Create router and setup routes
	router := routes.NewRouter(grpcServer)
	router.SetupRoutes()

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	return &Server{
		server:    grpcServer,
		container: container,
	}
}

// Start starts the gRPC server with graceful shutdown
func (s *Server) Start() error {
	conf := s.container.GetConfig()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(conf.GrpcPort))
	if err != nil {
		slog.Error("Failed to listen", "error", err, "port", conf.GrpcPort)
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create a channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests in a separate goroutine.
	go func() {
		slog.Info("Starting gRPC server", "port", conf.GrpcPort)
		fmt.Printf("Starting gRPC server on port %d\n", conf.GrpcPort)
		serverErrors <- s.server.Serve(listener)
	}()

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("error starting gRPC server: %w", err)
	case <-s.shutdown():
		slog.Info("Start gRPC shutdown...")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Gracefully shutdown the server.
		done := make(chan struct{})
		go func() {
			s.server.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			slog.Info("gRPC server stopped gracefully")
		case <-ctx.Done():
			slog.Warn("Graceful shutdown timeout, forcing stop")
			s.server.Stop()
		}
	}

	// Cleanup container connections
	if err := s.container.Close(); err != nil {
		slog.Error("Failed to close container connections", "error", err)
	} else {
		slog.Info("Container connections closed successfully")
	}

	return nil
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() error {
	slog.Info("Stopping gRPC server gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		// Cleanup container connections
		if err := s.container.Close(); err != nil {
			slog.Error("Failed to close container connections", "error", err)
		} else {
			slog.Info("Container connections closed successfully")
		}
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return fmt.Errorf("graceful shutdown timeout")
	}
}

// GetCache returns the cache instance
func (s *Server) GetCache() *redis.Client {
	return s.container.GetRedis()
}

// shutdown returns a channel that will be closed when shutdown should occur.
func (s *Server) shutdown() chan os.Signal {
	// Make channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	return shutdown
}
