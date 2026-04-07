package config

import (
	"fmt"
)

// Mode represents the application mode
type Mode string

const (
	Development Mode = "development"
	Production  Mode = "production"
	Testing     Mode = "testing"
)

// PostgresDatabase represents PostgreSQL database configuration
type PostgresDatabase struct {
	Host                string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Database            string `envconfig:"POSTGRES_DATABASE"`
	User                string `envconfig:"POSTGRES_USER"`
	Password            string `envconfig:"POSTGRES_PASSWORD"`
	Port                int    `envconfig:"POSTGRES_PORT" default:"5432"`
	ReplicaSet          string `envconfig:"POSTGRES_REPLICA_SET"`
	SSL                 bool   `envconfig:"POSTGRES_SSL" default:"false"`
	MaxPoolSize         uint64 `envconfig:"POSTGRES_MAX_POOL_SIZE" default:"10"`
	MinPoolSize         uint64 `envconfig:"POSTGRES_MIN_POOL_SIZE" default:"1"`
	MaxConnIdleTimeInMs int    `envconfig:"POSTGRES_MAX_CONN_IDLE_TIME_MS" default:"300000"`
}

// RedisConfig represents Redis cache configuration
type RedisConfig struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     int    `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD"`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
	PoolSize int    `envconfig:"REDIS_POOL_SIZE" default:"10"`
	MinIdle  int    `envconfig:"REDIS_MIN_IDLE" default:"5"`
}

type CorsOrigin struct {
	Origin []string `envconfig:"CORS_ORIGIN"`
}

// Config represents the application configuration
type Config struct {
	Version          string `envconfig:"VERSION" default:"1.0.0"`
	Mode             Mode   `envconfig:"MODE" default:"development"`
	ServiceName      string `envconfig:"SERVICE_NAME" default:"starter-service"`
	HttpPort         int    `envconfig:"HTTP_PORT" default:"8080"`
	HealthCheckRoute string `envconfig:"HEALTH_CHECK_ROUTE" default:"/health"`
	ApiVersion       string `envconfig:"API_VERSION" default:"v1"`
	JwtSecret        string `envconfig:"JWT_SECRET" default:"secret"`
	ServiceBasePath  string `envconfig:"SERVICE_BASE_PATH" default:"/api/v1"`

	// Database configuration
	Database PostgresDatabase

	// Cache configuration
	Redis RedisConfig

	// Logging
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// CORS configuration
	Cors CorsOrigin
}

var config *Config

// GetConfig returns the application configuration
func GetConfig() *Config {
	if config == nil {
		// Load configuration from .env file and environment variables
		if err := LoadConfig(); err != nil {
			panic(err)
		}
	}
	return config
}

// IsDevelopment checks if the application is in development mode
func (c *Config) IsDevelopment() bool {
	return c.Mode == Development
}

// IsProduction checks if the application is in production mode
func (c *Config) IsProduction() bool {
	return c.Mode == Production
}

// IsTesting checks if the application is in testing mode
func (c *Config) IsTesting() bool {
	return c.Mode == Testing
}

// GetDatabaseURL returns the database URL
func (c *Config) GetDatabaseURL() string {
	if c.Database.User != "" && c.Database.Password != "" && c.Database.Database != "" {
		sslMode := "disable"
		if c.Database.SSL {
			sslMode = "require"
		}
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Database, sslMode)
	}

	return ""
}
