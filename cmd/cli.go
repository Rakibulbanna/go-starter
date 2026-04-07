package cmd

import (
	"skoolz/internal/interfaces/cli"

	"github.com/spf13/cobra"
)

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Interactive CLI commands",
	Long:  "Run interactive CLI commands for service management",
	RunE:  runCLI,
}

func runCLI(cmd *cobra.Command, args []string) error {
	cliApp := cli.NewCLI()
	return cliApp.Execute()
}

func init() {
	RootCmd.AddCommand(cliCmd)
}
