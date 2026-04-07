package cache

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"skoolz/config"
	AppError "skoolz/internal/shared/error"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	config *config.RedisConfig
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg *config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdle,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		config: cfg,
	}, nil
}

// NewRedisClient creates a new Redis client using the application config
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

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// Get retrieves a value from Redis by key
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return result, nil
}

// Set stores a key-value pair in Redis with optional expiration
func (r *RedisCache) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return result > 0, nil
}

// Expire sets expiration time for a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return nil
}

// TTL returns the remaining time to live for a key
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return result, nil
}

// Flush removes all keys from Redis
func (r *RedisCache) Flush(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush Redis: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Ping checks if Redis is available
func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return AppError.NewAPIErrorWithError(http.StatusInternalServerError, "REDIS_ERROR", "Redis ping failed", err)
	}
	return nil
}

// GetClient returns the underlying Redis client for advanced operations
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}
