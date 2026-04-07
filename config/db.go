package config

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewPostgresDB creates and returns a new sqlx.DB connection using the Postgres config.
func NewPostgresDB() (*sqlx.DB, error) {
	cfg := GetConfig()

	// Use the GetDatabaseURL method which handles SSL configuration properly
	connStr := cfg.GetDatabaseURL()
	if connStr == "" {
		return nil, fmt.Errorf("no database configuration found")
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}
