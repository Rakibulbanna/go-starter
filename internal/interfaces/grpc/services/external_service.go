//go:build grpc

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"skoolz/config"
	"skoolz/internal/infrastructure/external"
	welcomepb "skoolz/proto/welcome"
)

// ExternalServiceExample demonstrates external service consumption
type ExternalServiceExample struct {
	welcomepb.UnimplementedWelcomeServiceServer
	serviceFactory *external.ServiceFactory
	serviceName    string
}

// NewExternalServiceExample creates a new external service example
func NewExternalServiceExample(serviceName string, serviceFactory *external.ServiceFactory) *ExternalServiceExample {
	return &ExternalServiceExample{
		serviceFactory: serviceFactory,
		serviceName:    serviceName,
	}
}

// WelcomeWithExternalData demonstrates consuming external HTTP service
func (s *ExternalServiceExample) Welcome(ctx context.Context, req *welcomepb.WelcomeRequest) (*welcomepb.WelcomeResponse, error) {
	// Check if context is cancelled
	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	conf := config.GetConfig()

	// Example: Consume external HTTP service
	externalData, err := s.fetchExternalData(ctx)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to fetch external data: %v\n", err)
		externalData = "External service unavailable"
	}

	response := &welcomepb.WelcomeResponse{
		Message:        fmt.Sprintf("Welcome to the %s API! External data: %s", s.serviceName, externalData),
		Service:        s.serviceName,
		Timestamp:      timestamppb.Now(),
		ProjectVersion: conf.Version,
	}

	return response, nil
}

// fetchExternalData demonstrates how to consume an external HTTP service
func (s *ExternalServiceExample) fetchExternalData(ctx context.Context) (string, error) {
	// Get HTTP client for external service
	httpClient, err := s.serviceFactory.GetHTTPClient("example-api")
	if err != nil {
		return "", fmt.Errorf("failed to get HTTP client: %w", err)
	}

	// Make request to external service
	resp, err := httpClient.Get(ctx, "/api/data", map[string]string{
		"X-Custom-Header": "skoolz",
	})
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response (example)
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract relevant data
	if message, ok := data["message"].(string); ok {
		return message, nil
	}

	return "No message found", nil
}

// ExampleGrpcCall demonstrates how to consume an external gRPC service
func (s *ExternalServiceExample) ExampleGrpcCall(ctx context.Context) (string, error) {
	// Get gRPC client for external service
	grpcClient, err := s.serviceFactory.GetGrpcClient("example-grpc-service")
	if err != nil {
		return "", fmt.Errorf("failed to get gRPC client: %w", err)
	}

	// Create context with metadata
	_ = grpcClient.CreateContextWithMetadata(ctx, map[string]string{
		"service-name": s.serviceName,
		"request-id":   "example-request",
	})

	// Example: Call external gRPC service
	// Note: You would need to import the specific proto generated client
	// conn := grpcClient.GetConnection()
	// client := examplepb.NewExampleServiceClient(conn)
	// response, err := client.SomeMethod(ctxWithMetadata, &examplepb.SomeRequest{})

	// For now, return a placeholder
	return "gRPC call would be made here", nil
}
