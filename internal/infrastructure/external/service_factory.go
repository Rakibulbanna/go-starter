package external

import (
	"fmt"
	"sync"

	"skoolz/config"
)

// ServiceFactory manages external service clients
type ServiceFactory struct {
	httpClients map[string]*HTTPClient
	grpcClients map[string]*GrpcClient
	mutex       sync.RWMutex
}

// NewServiceFactory creates a new service factory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{
		httpClients: make(map[string]*HTTPClient),
		grpcClients: make(map[string]*GrpcClient),
	}
}

// GetHTTPClient returns an HTTP client for the specified service
func (sf *ServiceFactory) GetHTTPClient(serviceName string) (*HTTPClient, error) {
	sf.mutex.RLock()
	client, exists := sf.httpClients[serviceName]
	sf.mutex.RUnlock()

	if exists {
		return client, nil
	}

	// Create new client
	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	// Double-check after acquiring write lock
	if client, exists = sf.httpClients[serviceName]; exists {
		return client, nil
	}

	// Get configuration for the service
	conf := config.GetConfig()
	serviceConfig, exists := conf.ExternalServices[serviceName]
	if !exists {
		return nil, fmt.Errorf("service configuration not found for: %s", serviceName)
	}

	// Create new HTTP client
	client = NewHTTPClient(serviceConfig)
	sf.httpClients[serviceName] = client

	return client, nil
}

// GetGrpcClient returns a gRPC client for the specified service
func (sf *ServiceFactory) GetGrpcClient(serviceName string) (*GrpcClient, error) {
	sf.mutex.RLock()
	client, exists := sf.grpcClients[serviceName]
	sf.mutex.RUnlock()

	if exists {
		return client, nil
	}

	// Create new client
	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	// Double-check after acquiring write lock
	if client, exists = sf.grpcClients[serviceName]; exists {
		return client, nil
	}

	// Get configuration for the service
	conf := config.GetConfig()
	serviceConfig, exists := conf.GrpcServices[serviceName]
	if !exists {
		return nil, fmt.Errorf("gRPC service configuration not found for: %s", serviceName)
	}

	// Create new gRPC client
	client, err := NewGrpcClient(serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", serviceName, err)
	}

	sf.grpcClients[serviceName] = client

	return client, nil
}

// CloseAll closes all external service connections
func (sf *ServiceFactory) CloseAll() error {
	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	// Close all gRPC connections
	for name, client := range sf.grpcClients {
		if err := client.Close(); err != nil {
			return fmt.Errorf("failed to close gRPC client %s: %w", name, err)
		}
	}

	// Clear the maps
	sf.httpClients = make(map[string]*HTTPClient)
	sf.grpcClients = make(map[string]*GrpcClient)

	return nil
}
