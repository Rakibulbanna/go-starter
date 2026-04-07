//go:build grpc

package routes

import (
	"skoolz/config"
	"skoolz/internal/infrastructure/external"
	"skoolz/internal/interfaces/grpc/services"
	healthpb "skoolz/proto/health"
	welcomepb "skoolz/proto/welcome"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Router handles gRPC service registration
type Router struct {
	server         *grpc.Server
	serviceFactory *external.ServiceFactory
}

// NewRouter creates a new gRPC router
func NewRouter(server *grpc.Server) *Router {
	return &Router{
		server:         server,
		serviceFactory: external.NewServiceFactory(),
	}
}

// SetupRoutes configures all gRPC services
func (r *Router) SetupRoutes() {
	conf := config.GetConfig()

	// Create services
	healthService := services.NewHealthService(conf.ServiceName, conf.Version)
	welcomeService := services.NewWelcomeService(conf.ServiceName)

	// Create external service example (demonstrates external service consumption)
	_ = services.NewExternalServiceExample(conf.ServiceName, r.serviceFactory)

	// Register services
	healthpb.RegisterHealthServiceServer(r.server, healthService)
	welcomepb.RegisterWelcomeServiceServer(r.server, welcomeService)

	// Register external service example (you can create your own proto for this)
	// externalpb.RegisterExternalServiceServer(r.server, externalServiceExample)

	// Enable reflection for debugging
	reflection.Register(r.server)
}

// GetServer returns the underlying gRPC server
func (r *Router) GetServer() *grpc.Server {
	return r.server
}

// GetServiceFactory returns the service factory for external services
func (r *Router) GetServiceFactory() *external.ServiceFactory {
	return r.serviceFactory
}

// Close closes all external service connections
func (r *Router) Close() error {
	return r.serviceFactory.CloseAll()
}
