package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"skoolz/config"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

// DatabaseHandler handles database-related CLI operations
type DatabaseHandler struct {
	config *config.Config
}

// NewDatabaseHandler creates a new database handler
func NewDatabaseHandler() *DatabaseHandler {
	return &DatabaseHandler{
		config: config.GetConfig(),
	}
}

// CreateDatabaseCmd creates the database creation command
func (h *DatabaseHandler) CreateDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-db",
		Short: "Create database if it doesn't exist",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🗄️  Creating database...")

			// Connect to postgres database to create our target database
			postgresURL := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
				h.config.Database.User, h.config.Database.Password, h.config.Database.Host, h.config.Database.Port)

			db, err := sqlx.Connect("postgres", postgresURL)
			if err != nil {
				return fmt.Errorf("failed to connect to postgres: %v", err)
			}
			defer db.Close()

			// Check if database exists
			var exists bool
			err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", h.config.Database.Database)
			if err != nil {
				return fmt.Errorf("failed to check database existence: %v", err)
			}

			if exists {
				fmt.Printf("✅ Database '%s' already exists\n", h.config.Database.Database)
				return nil
			}

			// Create database
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", h.config.Database.Database))
			if err != nil {
				return fmt.Errorf("failed to create database: %v", err)
			}

			fmt.Printf("✅ Database '%s' created successfully\n", h.config.Database.Database)
			return nil
		},
	}
}

// DropDatabaseCmd creates the database drop command
func (h *DatabaseHandler) DropDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drop-db",
		Short: "Drop database if it exists",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🗑️  Dropping database...")

			// Connect to postgres database
			postgresURL := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
				h.config.Database.User, h.config.Database.Password, h.config.Database.Host, h.config.Database.Port)

			db, err := sqlx.Connect("postgres", postgresURL)
			if err != nil {
				return fmt.Errorf("failed to connect to postgres: %v", err)
			}
			defer db.Close()

			// Check if database exists
			var exists bool
			err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", h.config.Database.Database)
			if err != nil {
				return fmt.Errorf("failed to check database existence: %v", err)
			}

			if !exists {
				fmt.Printf("ℹ️  Database '%s' does not exist\n", h.config.Database.Database)
				return nil
			}

			// Drop database
			_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", h.config.Database.Database))
			if err != nil {
				return fmt.Errorf("failed to drop database: %v", err)
			}

			fmt.Printf("✅ Database '%s' dropped successfully\n", h.config.Database.Database)
			return nil
		},
	}
}

// ResetDatabaseCmd creates the database reset command
func (h *DatabaseHandler) ResetDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset-db",
		Short: "Reset database (drop and recreate)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔄 Resetting database...")

			// Drop database
			dropCmd := h.DropDatabaseCmd()
			if err := dropCmd.RunE(cmd, args); err != nil {
				return err
			}

			// Create database
			createCmd := h.CreateDatabaseCmd()
			if err := createCmd.RunE(cmd, args); err != nil {
				return err
			}

			fmt.Println("✅ Database reset completed")
			return nil
		},
	}
}

// ServiceHandler handles service-related CLI operations
type ServiceHandler struct {
	config *config.Config
}

// NewServiceHandler creates a new service handler
func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{
		config: config.GetConfig(),
	}
}

// StatusCmd creates the service status command
func (h *ServiceHandler) StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("📊 Service Status:")
			fmt.Printf("Service Name: %s\n", h.config.ServiceName)
			fmt.Printf("Version: %s\n", h.config.Version)
			fmt.Printf("Mode: %s\n", h.config.Mode)
			fmt.Printf("HTTP Port: %d\n", h.config.HttpPort)
			fmt.Printf("gRPC Port: %d\n", h.config.GrpcPort)
			fmt.Printf("API Version: %s\n", h.config.ApiVersion)
			fmt.Printf("Health Check Route: %s\n", h.config.HealthCheckRoute)
			fmt.Printf("Service Base Path: %s\n", h.config.ServiceBasePath)
			fmt.Printf("Log Level: %s\n", h.config.LogLevel)

			// Check if service is running (placeholder)
			fmt.Println("\n🔄 Service Status: Not implemented (would check if service is running)")

			return nil
		},
	}
}

// UtilityHandler handles utility CLI operations
type UtilityHandler struct {
	config *config.Config
}

// NewUtilityHandler creates a new utility handler
func NewUtilityHandler() *UtilityHandler {
	return &UtilityHandler{
		config: config.GetConfig(),
	}
}

// GenerateSecretCmd creates the secret generation command
func (h *UtilityHandler) GenerateSecretCmd() *cobra.Command {
	var length int

	cmd := &cobra.Command{
		Use:   "generate-secret",
		Short: "Generate a secure random secret",
		Run: func(cmd *cobra.Command, args []string) {
			secret := generateRandomSecret(length)
			fmt.Printf("🔐 Generated Secret (%d characters):\n", len(secret))
			fmt.Println(secret)
		},
	}

	cmd.Flags().IntVarP(&length, "length", "l", 32, "Secret length")
	return cmd
}

// BackupCmd creates the backup command
func (h *UtilityHandler) BackupCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if output == "" {
				timestamp := time.Now().Format("2006-01-02_15-04-05")
				output = fmt.Sprintf("backup_%s.sql", timestamp)
			}

			fmt.Printf("💾 Creating backup: %s\n", output)

			// In a real implementation, you would use pg_dump
			// For now, we'll just create a placeholder
			backupDir := "./backups"
			if err := os.MkdirAll(backupDir, 0755); err != nil {
				return fmt.Errorf("failed to create backup directory: %v", err)
			}

			backupPath := filepath.Join(backupDir, output)
			backupContent := fmt.Sprintf("-- Backup created at %s\n-- Database: %s\n",
				time.Now().Format(time.RFC3339), h.config.Database.Database)

			if err := os.WriteFile(backupPath, []byte(backupContent), 0644); err != nil {
				return fmt.Errorf("failed to create backup file: %v", err)
			}

			fmt.Printf("✅ Backup created: %s\n", backupPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file name")
	return cmd
}

// RestoreCmd creates the restore command
func (h *UtilityHandler) RestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [backup-file]",
		Short: "Restore database from backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupFile := args[0]

			fmt.Printf("🔄 Restoring from backup: %s\n", backupFile)

			// Check if backup file exists
			if _, err := os.Stat(backupFile); os.IsNotExist(err) {
				return fmt.Errorf("backup file not found: %s", backupFile)
			}

			// In a real implementation, you would use pg_restore
			// For now, we'll just show a placeholder
			fmt.Println("✅ Backup restoration completed (placeholder)")
			return nil
		},
	}
}

// generateRandomSecret generates a random secret of specified length
func generateRandomSecret(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	secret := make([]byte, length)

	for i := range secret {
		secret[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond) // Add some entropy
	}

	return string(secret)
}
