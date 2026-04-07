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

	if err := godotenv.Load(); err != nil {
		slog.Warn(".env not found, that's okay!")
	}

	viper.AutomaticEnv()

	config = &Config{
		Version:          viper.GetString("VERSION"),
		Mode:             Mode(viper.GetString("MODE")),
		ServiceName:      viper.GetString("SERVICE_NAME"),
		HttpPort:         viper.GetInt("HTTP_PORT"),
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
	if err := v.Struct(config); err != nil {
		exit(err)
	}

	return nil
}
