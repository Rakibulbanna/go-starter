package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"skoolz/internal/infrastructure/container"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE:  migrate,
}

func migrate(cmd *cobra.Command, args []string) error {
	// Initialize container
	container := container.GetContainer()
	if err := container.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize container: %v", err)
	}
	defer container.Close()

	db := container.GetDB()
	logger := container.GetLogger()

	// Get migrations directory
	migrationsDir := "./internal/infrastructure/database/postgres/migrations"

	// Read all migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Filter and sort migration files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	logger.Info("Starting database migrations", "count", len(migrationFiles))

	// Create migrations table if it doesn't exist
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	_, err = db.Exec(createMigrationsTable)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Execute each migration
	for _, filename := range migrationFiles {
		// Check if migration already executed
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM migrations WHERE filename = $1", filename)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		if count > 0 {
			logger.Info("Migration already executed", "filename", filename)
			continue
		}

		// Read and execute migration
		migrationPath := filepath.Join(migrationsDir, filename)
		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", filename, err)
		}

		// Execute migration in transaction
		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}

		_, err = tx.Exec(string(migrationSQL))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %v", filename, err)
		}

		// Record migration execution
		_, err = tx.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %v", filename, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit migration %s: %v", filename, err)
		}

		logger.Info("Migration executed successfully", "filename", filename)
	}

	logger.Info("All migrations completed successfully")
	return nil
}
