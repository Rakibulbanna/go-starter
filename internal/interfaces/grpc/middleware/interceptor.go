//go:build grpc

package middleware

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

// UnaryInterceptor provides logging and error handling for gRPC calls
func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	slog.Info("gRPC request", "method", info.FullMethod)

	resp, err := handler(ctx, req)
	if err != nil {
		slog.Error("gRPC request failed", "method", info.FullMethod, "error", err)
	} else {
		slog.Info("gRPC request completed", "method", info.FullMethod)
	}

	return resp, err
}
