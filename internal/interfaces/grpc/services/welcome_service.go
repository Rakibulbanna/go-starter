//go:build grpc

package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"skoolz/config"
	welcomepb "skoolz/proto/welcome"
)

// WelcomeService handles gRPC welcome requests
type WelcomeService struct {
	welcomepb.UnimplementedWelcomeServiceServer
	serviceName string
}

// NewWelcomeService creates a new welcome service
func NewWelcomeService(serviceName string) *WelcomeService {
	return &WelcomeService{
		serviceName: serviceName,
	}
}

// Welcome handles the welcome gRPC endpoint
func (s *WelcomeService) Welcome(ctx context.Context, req *welcomepb.WelcomeRequest) (*welcomepb.WelcomeResponse, error) {
	// Check if context is cancelled
	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	conf := config.GetConfig()
	response := &welcomepb.WelcomeResponse{
		Message:        "Welcome to the " + s.serviceName + " API!",
		Service:        s.serviceName,
		Timestamp:      timestamppb.Now(),
		ProjectVersion: conf.Version,
	}

	return response, nil
}
