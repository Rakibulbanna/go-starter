package cmd

import (
	"skoolz/internal/infrastructure/container"
	"skoolz/internal/interfaces/http"

	"github.com/spf13/cobra"
)

var serveRestCmd = &cobra.Command{
	Use:   "serve-rest",
	Short: "Serve the REST API",
	RunE:  serveRest,
}

func serveRest(cmd *cobra.Command, args []string) error {
	// Initialize container with database connection
	container := container.GetContainer()
	if err := container.Initialize(); err != nil {
		return err
	}
	defer container.Close()

	server := http.NewServer()
	return server.Start()
}
