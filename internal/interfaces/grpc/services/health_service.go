//go:build grpc

package services

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	health "skoolz/proto/health"
)

// HealthService handles gRPC health check requests
type HealthService struct {
	health.UnimplementedHealthServiceServer
	serviceName string
	version     string
}

// NewHealthService creates a new health service
func NewHealthService(serviceName, version string) *HealthService {
	return &HealthService{
		serviceName: serviceName,
		version:     version,
	}
}

// HealthCheck handles the health check gRPC endpoint
func (s *HealthService) HealthCheck(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	// Check if context is cancelled
	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	response := &health.HealthCheckResponse{
		Status:    "ok",
		Service:   s.serviceName,
		Timestamp: timestamppb.New(time.Now()),
		Version:   s.version,
	}

	return response, nil
}
