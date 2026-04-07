package cache

import (
	"context"
	"fmt"
	"time"

	"skoolz/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client using the application config.
// Returns an error if Redis is unreachable; the caller decides whether that
// is fatal (the REST API treats it as a non-fatal warning).
func NewRedisClient() (*redis.Client, error) {
	cfg := config.GetConfig()

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdle,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
