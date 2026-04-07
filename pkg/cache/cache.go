package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string) (string, error)

	// Set stores a key-value pair in cache with optional expiration
	Set(ctx context.Context, key, value string, expiration time.Duration) error

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Expire sets expiration time for a key
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// TTL returns the remaining time to live for a key
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Flush removes all keys from cache
	Flush(ctx context.Context) error

	// Close closes the cache connection
	Close() error

	// Ping checks if the cache is available
	Ping(ctx context.Context) error
}
