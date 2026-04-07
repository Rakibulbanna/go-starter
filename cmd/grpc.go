//go:build grpc

package cmd

import (
	"skoolz/internal/infrastructure/container"
	"skoolz/internal/interfaces/grpc"

	"github.com/spf13/cobra"
)

var serveGrpcCmd = &cobra.Command{
	Use:   "serve-grpc",
	Short: "Serve the gRPC API",
	RunE:  serveGrpc,
}

func registerGrpcCmd() {
	RootCmd.AddCommand(serveGrpcCmd)
}

func serveGrpc(cmd *cobra.Command, args []string) error {
	// Initialize container with database connection
	container := container.GetContainer()
	if err := container.Initialize(); err != nil {
		return err
	}
	defer container.Close()

	server := grpc.NewServer()
	return server.Start()
}
