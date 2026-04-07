package config

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadConfig() error {
	exit := func(err error) {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env not found, that's okay!")
	}

	viper.AutomaticEnv()

	config = &Config{
		Version:          viper.GetString("VERSION"),
		Mode:             Mode(viper.GetString("MODE")),
		ServiceName:      viper.GetString("SERVICE_NAME"),
		HttpPort:         viper.GetInt("HTTP_PORT"),
		GrpcPort:         viper.GetInt("GRPC_PORT"),
		HealthCheckRoute: viper.GetString("HEALTH_CHECK_ROUTE"),
		ApiVersion:       viper.GetString("API_VERSION"),
		JwtSecret:        viper.GetString("JWT_SECRET"),
		ServiceBasePath:  viper.GetString("SERVICE_BASE_PATH"),

		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
			PoolSize: viper.GetInt("REDIS_POOL_SIZE"),
			MinIdle:  viper.GetInt("REDIS_MIN_IDLE"),
		},

		Kafka: KafkaConfig{
			Brokers:     []string{viper.GetString("KAFKA_BROKERS")},
			ClientID:    viper.GetString("KAFKA_CLIENT_ID"),
			GroupID:     viper.GetString("KAFKA_GROUP_ID"),
			EnableKafka: viper.GetBool("KAFKA_ENABLE"),
			Topics: struct {
				UserEvents   string `envconfig:"KAFKA_TOPIC_EVENTS" default:"user-events"`
				UserCommands string `envconfig:"KAFKA_TOPIC_COMMANDS" default:"user-commands"`
			}{
				UserEvents:   viper.GetString("KAFKA_TOPIC_EVENTS"),
				UserCommands: viper.GetString("KAFKA_TOPIC_COMMANDS"),
			},
			Consumer: struct {
				AutoOffsetReset string `envconfig:"KAFKA_CONSUMER_AUTO_OFFSET_RESET" default:"earliest"`
				MaxBytes        int    `envconfig:"KAFKA_CONSUMER_MAX_BYTES" default:"1048576"`
				MaxWaitSeconds  int    `envconfig:"KAFKA_CONSUMER_MAX_WAIT_SECONDS" default:"10"`
			}{
				AutoOffsetReset: viper.GetString("KAFKA_CONSUMER_AUTO_OFFSET_RESET"),
				MaxBytes:        viper.GetInt("KAFKA_CONSUMER_MAX_BYTES"),
				MaxWaitSeconds:  viper.GetInt("KAFKA_CONSUMER_MAX_WAIT_SECONDS"),
			},
			Producer: struct {
				RequiredAcks int `envconfig:"KAFKA_PRODUCER_REQUIRED_ACKS" default:"1"`
				MaxAttempts  int `envconfig:"KAFKA_PRODUCER_MAX_ATTEMPTS" default:"3"`
			}{
				RequiredAcks: viper.GetInt("KAFKA_PRODUCER_REQUIRED_ACKS"),
				MaxAttempts:  viper.GetInt("KAFKA_PRODUCER_MAX_ATTEMPTS"),
			},
		},

		Nats: NatsConfig{
			URL:      viper.GetString("NATS_URL"),
			ClientID: viper.GetString("NATS_CLIENT_ID"),
			Cluster:  viper.GetString("NATS_CLUSTER"),
			Subjects: struct {
				UserEvents   string `envconfig:"NATS_SUBJECT_EVENTS" default:"user.events"`
				UserCommands string `envconfig:"NATS_SUBJECT_COMMANDS" default:"user.commands"`
				UserQueries  string `envconfig:"NATS_SUBJECT_QUERIES" default:"user.queries"`
			}{
				UserEvents:   viper.GetString("NATS_SUBJECT_EVENTS"),
				UserCommands: viper.GetString("NATS_SUBJECT_COMMANDS"),
				UserQueries:  viper.GetString("NATS_SUBJECT_QUERIES"),
			},
			QueueGroup: viper.GetString("NATS_QUEUE_GROUP"),
		},

		ExternalServices: map[string]ExternalService{
			"external-service": {
				Name:    viper.GetString("EXTERNAL_SERVICE_NAME"),
				BaseURL: viper.GetString("EXTERNAL_SERVICE_BASE_URL"),
				Timeout: viper.GetInt("EXTERNAL_SERVICE_TIMEOUT"),
				APIKey:  viper.GetString("EXTERNAL_SERVICE_API_KEY"),
			}},

		GrpcServices: map[string]GrpcService{
			"grpc-service": {
				Host:    viper.GetString("GRPC_SERVICE_HOST"),
				Port:    viper.GetInt("GRPC_SERVICE_PORT"),
				Timeout: viper.GetInt("GRPC_SERVICE_TIMEOUT"),
			},
		},

		Database: PostgresDatabase{
			Host:                viper.GetString("POSTGRES_HOST"),
			Database:            viper.GetString("POSTGRES_DATABASE"),
			User:                viper.GetString("POSTGRES_USER"),
			Password:            viper.GetString("POSTGRES_PASSWORD"),
			Port:                viper.GetInt("POSTGRES_PORT"),
			ReplicaSet:          viper.GetString("POSTGRES_REPLICA_SET"),
			SSL:                 viper.GetBool("POSTGRES_SSL"),
			MaxPoolSize:         viper.GetUint64("POSTGRES_MAX_POOL_SIZE"),
			MinPoolSize:         viper.GetUint64("POSTGRES_MIN_POOL_SIZE"),
			MaxConnIdleTimeInMs: viper.GetInt("POSTGRES_MAX_CONN_IDLE_TIME_MS"),
		},

		Cors: CorsOrigin{
			Origin: viper.GetStringSlice("CORS_ORIGIN"),
		},
	}

	v := validator.New()
	if err = v.Struct(config); err != nil {
		exit(err)
	}

	return nil
}
