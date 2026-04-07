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

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers     []string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	ClientID    string   `envconfig:"KAFKA_CLIENT_ID" default:"skoolz"`
	GroupID     string   `envconfig:"KAFKA_GROUP_ID" default:"skoolz-group"`
	EnableKafka bool     `envconfig:"KAFKA_ENABLE" default:"true"`
	Topics      struct {
		UserEvents   string `envconfig:"KAFKA_TOPIC_EVENTS" default:"user-events"`
		UserCommands string `envconfig:"KAFKA_TOPIC_COMMANDS" default:"user-commands"`
	}
	Consumer struct {
		AutoOffsetReset string `envconfig:"KAFKA_CONSUMER_AUTO_OFFSET_RESET" default:"earliest"`
		MaxBytes        int    `envconfig:"KAFKA_CONSUMER_MAX_BYTES" default:"1048576"`
		MaxWaitSeconds  int    `envconfig:"KAFKA_CONSUMER_MAX_WAIT_SECONDS" default:"10"`
	}
	Producer struct {
		RequiredAcks int `envconfig:"KAFKA_PRODUCER_REQUIRED_ACKS" default:"1"`
		MaxAttempts  int `envconfig:"KAFKA_PRODUCER_MAX_ATTEMPTS" default:"3"`
	}
}

// NatsConfig represents NATS configuration
type NatsConfig struct {
	URL      string `envconfig:"NATS_URL" default:"nats://localhost:4222"`
	ClientID string `envconfig:"NATS_CLIENT_ID" default:"skoolz"`
	Cluster  string `envconfig:"NATS_CLUSTER" default:"test-cluster"`
	Subjects struct {
		UserEvents   string `envconfig:"NATS_SUBJECT_EVENTS" default:"user.events"`
		UserCommands string `envconfig:"NATS_SUBJECT_COMMANDS" default:"user.commands"`
		UserQueries  string `envconfig:"NATS_SUBJECT_QUERIES" default:"user.queries"`
	}
	QueueGroup string `envconfig:"NATS_QUEUE_GROUP" default:"skoolz-group"`
}

// ExternalService represents external service configuration
type ExternalService struct {
	Name     string `envconfig:"EXTERNAL_SERVICE_NAME"`
	BaseURL  string `envconfig:"EXTERNAL_SERVICE_BASE_URL"`
	Timeout  int    `envconfig:"EXTERNAL_SERVICE_TIMEOUT" default:"30"`
	APIKey   string `envconfig:"EXTERNAL_SERVICE_API_KEY"`
	Username string `envconfig:"EXTERNAL_SERVICE_USERNAME"`
	Password string `envconfig:"EXTERNAL_SERVICE_PASSWORD"`
}

// GrpcService represents gRPC external service configuration
type GrpcService struct {
	Name    string `envconfig:"GRPC_SERVICE_NAME"`
	Host    string `envconfig:"GRPC_SERVICE_HOST"`
	Port    int    `envconfig:"GRPC_SERVICE_PORT" default:"50051"`
	Timeout int    `envconfig:"GRPC_SERVICE_TIMEOUT" default:"30"`
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
	GrpcPort         int    `envconfig:"GRPC_PORT" default:"8081"`
	HealthCheckRoute string `envconfig:"HEALTH_CHECK_ROUTE" default:"/health"`
	ApiVersion       string `envconfig:"API_VERSION" default:"v1"`
	JwtSecret        string `envconfig:"JWT_SECRET" default:"secret"`
	ServiceBasePath  string `envconfig:"SERVICE_BASE_PATH" default:"/api/v1"`

	// Database configuration
	Database PostgresDatabase

	// Cache configuration
	Redis RedisConfig

	// Message broker configuration
	Kafka KafkaConfig
	Nats  NatsConfig

	// External services configuration
	ExternalServices map[string]ExternalService
	GrpcServices     map[string]GrpcService

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
	// Use the new Database configuration if available
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
