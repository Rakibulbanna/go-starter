package container

import (
	"context"
	"fmt"
	"skoolz/config"
	"skoolz/pkg/cache"
	"skoolz/internal/infrastructure/messaging"
	"skoolz/internal/logger"
	"sync"

	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Container holds all application dependencies
type Container struct {
	db               *sqlx.DB
	redis            *redis.Client
	log              *slog.Logger
	messagingFactory *messaging.MessagingFactory
	mu               sync.RWMutex
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

// GetRedis returns the Redis client
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

// GetMessagingFactory returns the messaging factory
func (c *Container) GetMessagingFactory() *messaging.MessagingFactory {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.messagingFactory
}

// GetConfig returns the application configuration
func (c *Container) GetConfig() *config.Config {
	return config.GetConfig()
}

// Lock locks the container for exclusive access
func (c *Container) Lock() {
	c.mu.Lock()
}

// Unlock unlocks the container
func (c *Container) Unlock() {
	c.mu.Unlock()
}

// RLock locks the container for read access
func (c *Container) RLock() {
	c.mu.RLock()
}

// RUnlock unlocks the container for read access
func (c *Container) RUnlock() {
	c.mu.RUnlock()
}

// Initialize initializes all dependencies
func (c *Container) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Load configuration first
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

	// Create startup logger for concise messages
	startupLog := logger.NewStartupLogger(log)

	// Initialize database
	db, err := config.NewPostgresDB()
	if err != nil {
		startupLog.Error("Database connection failed", "error", err)
		return err
	}
	c.db = db
	startupLog.Database("Database connected")

	// Initialize Redis
	redisClient, err := cache.NewRedisClient()
	if err != nil {
		startupLog.Warning("Redis connection failed, continuing without cache", "error", err)
	} else {
		c.redis = redisClient
		startupLog.Cache("Redis connected")
	}

	// Initialize messaging factory
	messagingFactory := messaging.NewMessagingFactory(cfg, c.log)
	c.messagingFactory = messagingFactory

	// Initialize NATS streams
	// if err := messagingFactory.InitializeStreams(); err != nil {
	// 	startupLog.Warning("NATS streams initialization failed", "error", err)
	// } else {
	// 	startupLog.Stream("NATS streams ready")
	// }

	// Start message subscriptions
	ctx := context.Background()
	if err := messagingFactory.StartSubscriptions(ctx); err != nil {
		fmt.Println("Message subscriptions failed", "error", err)
		// startupLog.Warning("Message subscriptions failed", "error", err)
	} else {
		fmt.Println("Message subscriptions active")
	}

	return nil
}

// Close closes all connections
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	if c.db != nil {
		if err := c.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.redis != nil {
		if err := c.redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.messagingFactory != nil {
		if err := c.messagingFactory.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0] // Return first error for simplicity
	}

	return nil
}
