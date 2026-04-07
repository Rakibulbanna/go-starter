package external

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"skoolz/config"
)

// GrpcClient represents a gRPC client for external services
type GrpcClient struct {
	conn   *grpc.ClientConn
	config config.GrpcService
}

// NewGrpcClient creates a new gRPC client for external services
func NewGrpcClient(serviceConfig config.GrpcService) (*GrpcClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to the gRPC server
	conn, err := grpc.DialContext(ctx,
		fmt.Sprintf("%s:%d", serviceConfig.Host, serviceConfig.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &GrpcClient{
		conn:   conn,
		config: serviceConfig,
	}, nil
}

// GetConnection returns the underlying gRPC connection
func (c *GrpcClient) GetConnection() *grpc.ClientConn {
	return c.conn
}

// Close closes the gRPC connection
func (c *GrpcClient) Close() error {
	return c.conn.Close()
}

// CreateContextWithMetadata creates a context with metadata for gRPC calls
func (c *GrpcClient) CreateContextWithMetadata(ctx context.Context, metadataMap map[string]string) context.Context {
	md := metadata.New(nil)
	for key, value := range metadataMap {
		md.Set(key, value)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// IsConnected checks if the gRPC connection is still valid
func (c *GrpcClient) IsConnected() bool {
	return c.conn.GetState().String() == "READY"
}
