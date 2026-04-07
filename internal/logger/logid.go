package logger

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"log/slog"

	"github.com/google/uuid"
)

type logIDKeyType struct{}

var LogIDKey = logIDKeyType{}

const logIDString = "log_id"

var (
	// Global counter for sequential log IDs
	sequentialCounter uint64
	counterMutex      sync.Mutex
)

// LogIDGenerator provides methods to generate different types of log IDs
type LogIDGenerator struct{}

// NewLogIDGenerator creates a new log ID generator
func NewLogIDGenerator() *LogIDGenerator {
	return &LogIDGenerator{}
}

// GenerateUUID generates a UUID-based log ID
func (g *LogIDGenerator) GenerateUUID() string {
	return uuid.New().String()
}

// GenerateSequential generates a sequential log ID with timestamp
func (g *LogIDGenerator) GenerateSequential() string {
	counterMutex.Lock()
	defer counterMutex.Unlock()

	sequentialCounter++
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("log_%s_%06d", timestamp, sequentialCounter)
}

// GenerateRandom generates a random log ID
func (g *LogIDGenerator) GenerateRandom() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("log_%x", bytes)
}

// GenerateTimestamped generates a timestamped log ID
func (g *LogIDGenerator) GenerateTimestamped() string {
	timestamp := time.Now().Format("20060102150405.000")
	return fmt.Sprintf("log_%s", timestamp)
}

// GenerateWithPrefix generates a log ID with a custom prefix
func (g *LogIDGenerator) GenerateWithPrefix(prefix string) string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	return fmt.Sprintf("%s_%s_%x", prefix, timestamp, randomBytes)
}

// GenerateFromContext extracts log ID from context or generates a new one
func (g *LogIDGenerator) GenerateFromContext(ctx context.Context) string {
	if logID, ok := ctx.Value(LogIDKey).(string); ok && logID != "" {
		return logID
	}
	return g.GenerateUUID()
}

// ContextWithLogID creates a new context with a log ID
func ContextWithLogID(ctx context.Context, logID string) context.Context {
	return context.WithValue(ctx, LogIDKey, logID)
}

// ContextWithNewLogID creates a new context with a newly generated log ID
func ContextWithNewLogID(ctx context.Context) context.Context {
	generator := NewLogIDGenerator()
	logID := generator.GenerateUUID()
	return context.WithValue(ctx, LogIDKey, logID)
}

// GetLogIDFromContext extracts log ID from context
func GetLogIDFromContext(ctx context.Context) string {
	if logID, ok := ctx.Value(LogIDKey).(string); ok {
		return logID
	}
	return ""
}

// LogID creates a slog.Attr for log ID
func LogID(id string) slog.Attr {
	return slog.String(logIDString, id)
}

// LogIDFromContext creates a slog.Attr for log ID from context
func LogIDFromContext(ctx context.Context) slog.Attr {
	logID := GetLogIDFromContext(ctx)
	if logID == "" {
		logID = NewLogIDGenerator().GenerateUUID()
	}
	return slog.String(logIDString, logID)
}

// Global convenience functions
var (
	// Default generator instance
	defaultGenerator = NewLogIDGenerator()
)

// GenerateLogID generates a UUID-based log ID using the default generator
func GenerateLogID() string {
	return defaultGenerator.GenerateUUID()
}

// GenerateSequentialLogID generates a sequential log ID using the default generator
func GenerateSequentialLogID() string {
	return defaultGenerator.GenerateSequential()
}

// GenerateRandomLogID generates a random log ID using the default generator
func GenerateRandomLogID() string {
	return defaultGenerator.GenerateRandom()
}

// GenerateTimestampedLogID generates a timestamped log ID using the default generator
func GenerateTimestampedLogID() string {
	return defaultGenerator.GenerateTimestamped()
}

// GenerateLogIDWithPrefix generates a log ID with prefix using the default generator
func GenerateLogIDWithPrefix(prefix string) string {
	return defaultGenerator.GenerateWithPrefix(prefix)
}
