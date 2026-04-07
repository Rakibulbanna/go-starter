package container

import (
	"fmt"
	"sync"

	"log/slog"

	"skoolz/config"
	"skoolz/internal/logger"
	"skoolz/pkg/cache"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Container holds all application dependencies
type Container struct {
	db    *sqlx.DB
	redis *redis.Client
	log   *slog.Logger
	mu    sync.RWMutex
}

var (
	container *Container
	once      sync.Once
)

// GetContainer returns the singleton container instance
func GetContainer() *Container {
	once.Do(func() {
		container = &Container{}
	})
	return container
}

// GetDB returns the database connection
func (c *Container) GetDB() *sqlx.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

// GetRedis returns the Redis client (may be nil if Redis is unreachable)
func (c *Container) GetRedis() *redis.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.redis
}

// GetLogger returns the logger instance
func (c *Container) GetLogger() *slog.Logger {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.log
}

// GetConfig returns the application configuration
func (c *Container) GetConfig() *config.Config {
	return config.GetConfig()
}

// Lock locks the container for exclusive access
func (c *Container) Lock()    { c.mu.Lock() }
func (c *Container) Unlock()  { c.mu.Unlock() }
func (c *Container) RLock()   { c.mu.RLock() }
func (c *Container) RUnlock() { c.mu.RUnlock() }

// Initialize initializes all dependencies
func (c *Container) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := c.GetConfig()
	if cfg == nil {
		return fmt.Errorf("failed to load configuration")
	}

	// Initialize logger
	log, err := logger.NewLogger()
	if err != nil {
		return err
	}
	c.log = log

	startupLog := logger.NewStartupLogger(log)

	// Initialize database (required)
	db, err := config.NewPostgresDB()
	if err != nil {
		startupLog.Error("Database connection failed", "error", err)
		return err
	}
	c.db = db
	startupLog.Database("Database connected")

	// Initialize Redis (optional — REST API works without it)
	redisClient, err := cache.NewRedisClient()
	if err != nil {
		startupLog.Warning("Redis connection failed, continuing without cache", "error", err)
	} else {
		c.redis = redisClient
		startupLog.Cache("Redis connected")
	}

	return nil
}

// Close closes all connections. Returns the first error encountered, if any.
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var firstErr error
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			firstErr = err
		}
	}
	if c.redis != nil {
		if err := c.redis.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
