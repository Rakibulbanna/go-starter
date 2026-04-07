package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "skoolz",
	Short: "Skoolz REST API service",
}

func init() {
	RootCmd.AddCommand(serveRestCmd)
	RootCmd.AddCommand(migrateCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
