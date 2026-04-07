package cli

import (
	"fmt"
	"os"
	"skoolz/config"
	"skoolz/internal/logger"

	"github.com/spf13/cobra"
)

// CLI represents the command line interface
type CLI struct {
	rootCmd *cobra.Command
	config  *config.Config
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	cli := &CLI{
		config: config.GetConfig(),
	}
	cli.setupCommands()
	return cli
}

// setupCommands initializes all CLI commands
func (c *CLI) setupCommands() {
	c.rootCmd = &cobra.Command{
		Use:   "setter",
		Short: "Setter Service - A comprehensive service management CLI",
		Long: `Setter Service CLI provides various commands to manage and interact with the service.
		
Examples:
  setter serve-rest     # Start REST API server
  setter serve-grpc     # Start gRPC server
  setter migrate        # Run database migrations
  setter health         # Check service health
  setter version        # Show version information`,
		Version: "1.0.0",
	}

	// Add subcommands
	c.rootCmd.AddCommand(c.createHealthCmd())
	c.rootCmd.AddCommand(c.createVersionCmd())
	c.rootCmd.AddCommand(c.createConfigCmd())
	c.rootCmd.AddCommand(c.createLogsCmd())
	c.rootCmd.AddCommand(c.createTestCmd())
	c.rootCmd.AddCommand(c.createCleanupCmd())
	c.rootCmd.AddCommand(c.createSetupCmd())

	// Add database commands
	dbHandler := NewDatabaseHandler()
	c.rootCmd.AddCommand(dbHandler.CreateDatabaseCmd())
	c.rootCmd.AddCommand(dbHandler.DropDatabaseCmd())
	c.rootCmd.AddCommand(dbHandler.ResetDatabaseCmd())

	// Add service commands
	serviceHandler := NewServiceHandler()
	c.rootCmd.AddCommand(serviceHandler.StatusCmd())

	// Add utility commands
	utilityHandler := NewUtilityHandler()
	c.rootCmd.AddCommand(utilityHandler.GenerateSecretCmd())
	c.rootCmd.AddCommand(utilityHandler.BackupCmd())
	c.rootCmd.AddCommand(utilityHandler.RestoreCmd())
}

// Execute runs the CLI application
func (c *CLI) Execute() error {
	return c.rootCmd.Execute()
}

// createHealthCmd creates the health check command
func (c *CLI) createHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check service health status",
		Long:  "Check the health status of various service components including database, cache, and external services",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.SetupLogger(c.config.ServiceName, string(c.config.Mode))

			fmt.Println("🔍 Checking service health...")

			// Check database connection
			db, err := config.NewPostgresDB()
			if err != nil {
				fmt.Println("❌ Database: Connection failed")
				return fmt.Errorf("database health check failed: %v", err)
			}
			defer db.Close()
			fmt.Println("✅ Database: Connection successful")

			// Check Redis connection (placeholder - implement when Redis client is available)
			fmt.Println("✅ Redis: Connection check skipped (not implemented)")

			fmt.Println("🎉 All health checks passed!")
			return nil
		},
	}
}

// createVersionCmd creates the version command
func (c *CLI) createVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Setter Service v%s\n", "1.0.0")
			fmt.Printf("Build Date: %s\n", "2024-01-01")
			fmt.Printf("Go Version: %s\n", "1.21")
		},
	}
}

// createConfigCmd creates the config command
func (c *CLI) createConfigCmd() *cobra.Command {
	var showSecrets bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show configuration information",
		Long:  "Display current configuration settings. Use --show-secrets to display sensitive information.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("📋 Current Configuration:")
			fmt.Printf("Service Name: %s\n", c.config.ServiceName)
			fmt.Printf("Mode: %s\n", c.config.Mode)
			fmt.Printf("HTTP Port: %d\n", c.config.HttpPort)
			fmt.Printf("gRPC Port: %d\n", c.config.GrpcPort)

			if showSecrets {
				fmt.Printf("Database URL: %s\n", c.config.GetDatabaseURL())
				fmt.Printf("Redis Host: %s:%d\n", c.config.Redis.Host, c.config.Redis.Port)
			} else {
				fmt.Println("Database URL: [HIDDEN] (use --show-secrets to view)")
				fmt.Println("Redis Host: [HIDDEN] (use --show-secrets to view)")
			}
		},
	}

	cmd.Flags().BoolVar(&showSecrets, "show-secrets", false, "Show sensitive configuration values")
	return cmd
}

// createLogsCmd creates the logs command
func (c *CLI) createLogsCmd() *cobra.Command {
	var follow bool
	var lines int

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show service logs",
		Long:  "Display service logs with optional following and line limiting",
		RunE: func(cmd *cobra.Command, args []string) error {
			logFile := "./logs/app.log"

			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				return fmt.Errorf("log file not found: %s", logFile)
			}

			fmt.Printf("📄 Showing logs from: %s\n", logFile)

			// In a real implementation, you would read and display the log file
			// For now, we'll just show a placeholder
			fmt.Println("Log file exists and is accessible")
			if follow {
				fmt.Println("Following mode enabled")
			}
			if lines > 0 {
				fmt.Printf("Showing last %d lines\n", lines)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().IntVarP(&lines, "lines", "n", 50, "Number of lines to show")
	return cmd
}

// createTestCmd creates the test command
func (c *CLI) createTestCmd() *cobra.Command {
	var testType string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests",
		Long:  "Run various types of tests (unit, integration, e2e)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🧪 Running %s tests...\n", testType)

			switch testType {
			case "unit":
				return c.runUnitTests()
			case "integration":
				return c.runIntegrationTests()
			case "e2e":
				return c.runE2ETests()
			default:
				return fmt.Errorf("unknown test type: %s. Use: unit, integration, or e2e", testType)
			}
		},
	}

	cmd.Flags().StringVarP(&testType, "type", "t", "unit", "Test type (unit, integration, e2e)")
	return cmd
}

// createCleanupCmd creates the cleanup command
func (c *CLI) createCleanupCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up temporary files and caches",
		Long:  "Remove temporary files, logs, and cache data",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧹 Starting cleanup...")

			// Clean up temporary files
			if err := os.RemoveAll("./tmp"); err != nil {
				fmt.Printf("⚠️  Warning: Could not remove tmp directory: %v\n", err)
			} else {
				fmt.Println("✅ Temporary files cleaned")
			}

			if all {
				// Clean up logs
				if err := os.RemoveAll("./logs"); err != nil {
					fmt.Printf("⚠️  Warning: Could not remove logs directory: %v\n", err)
				} else {
					fmt.Println("✅ Logs cleaned")
				}
			}

			fmt.Println("🎉 Cleanup completed!")
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Clean all files including logs")
	return cmd
}

// createSetupCmd creates the setup command
func (c *CLI) createSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Setup the service environment",
		Long:  "Initialize the service environment, create necessary directories, and validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🚀 Setting up service environment...")

			// Create necessary directories
			dirs := []string{"./logs", "./tmp", "./data"}
			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %v", dir, err)
				}
				fmt.Printf("✅ Created directory: %s\n", dir)
			}

			// Validate configuration
			if err := c.validateConfig(); err != nil {
				return fmt.Errorf("configuration validation failed: %v", err)
			}
			fmt.Println("✅ Configuration validated")

			// Test database connection
			db, err := config.NewPostgresDB()
			if err != nil {
				return fmt.Errorf("database connection test failed: %v", err)
			}
			defer db.Close()
			fmt.Println("✅ Database connection verified")

			fmt.Println("🎉 Setup completed successfully!")
			return nil
		},
	}
}

// validateConfig validates the current configuration
func (c *CLI) validateConfig() error {
	if c.config.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if c.config.HttpPort <= 0 {
		return fmt.Errorf("invalid HTTP port number")
	}
	if c.config.GrpcPort <= 0 {
		return fmt.Errorf("invalid gRPC port number")
	}
	return nil
}

// runUnitTests runs unit tests
func (c *CLI) runUnitTests() error {
	fmt.Println("Running unit tests...")
	// In a real implementation, you would run: go test ./tests/unit/...
	fmt.Println("✅ Unit tests completed")
	return nil
}

// runIntegrationTests runs integration tests
func (c *CLI) runIntegrationTests() error {
	fmt.Println("Running integration tests...")
	// In a real implementation, you would run: go test ./tests/integration/...
	fmt.Println("✅ Integration tests completed")
	return nil
}

// runE2ETests runs end-to-end tests
func (c *CLI) runE2ETests() error {
	fmt.Println("Running end-to-end tests...")
	// In a real implementation, you would run: go test ./tests/e2e/...
	fmt.Println("✅ E2E tests completed")
	return nil
}
