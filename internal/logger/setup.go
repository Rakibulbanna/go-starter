package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"skoolz/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

func replacer(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		return slog.Time(string(TimeKey), a.Value.Resolve().Time())

	case slog.MessageKey:
		return slog.String(string(MessageKey), a.Value.String())

	case slog.SourceKey:
		src := a.Value.Any().(*slog.Source)
		if src.Function == "" {
			return slog.Attr{}
		}
		base := path.Base(src.File)
		return slog.String(
			string(CallerKey),
			fmt.Sprintf("%s:%s:%d", base, src.Function, src.Line),
		)

	case slog.LevelKey:
		l := a.Value.Any().(slog.Level)
		return slog.String(string(LevelKey), levels[l])

	default:
		return a
	}
}

// SetupLogger sets up the logger with automatic log ID generation
func SetupLogger(serviceName string, mode string) {
	var multiWriter io.Writer

	// In production mode, only log to stdout
	// In debug mode, log to both stdout and file
	if mode == "release" {
		multiWriter = os.Stdout
	} else {
		// Create logs directory if it doesn't exist
		if err := os.MkdirAll("logs", 0755); err != nil {
			panic(fmt.Sprintf("failed to create logs directory: %v", err))
		}

		// Use current date for filename
		currentDate := time.Now().Format("2006-01-02")
		filename := fmt.Sprintf("logs/%s.log", currentDate)

		// Setup file rotation
		fileLogger := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    100, // megabytes
			MaxBackups: 30,  // keep 30 days of logs
			MaxAge:     30,  // days
			Compress:   true,
		}

		// Create multi-writer for both file and stdout
		multiWriter = io.MultiWriter(os.Stdout, fileLogger)
	}

	// Setup JSON handler with Grafana-friendly format and automatic log ID
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: replacer,
	})).With(
		string(ServiceKey), serviceName,
		logIDString, GenerateLogID(), // Automatically add log ID to all entries
	)

	slog.SetDefault(logger)
}

// SetupLoggerWithContext sets up the logger with context-aware log ID generation
func SetupLoggerWithContext(serviceName string, mode string, ctx context.Context) {
	var multiWriter io.Writer

	// In production mode, only log to stdout
	// In debug mode, log to both stdout and file
	if mode == "release" {
		multiWriter = os.Stdout
	} else {
		// Create logs directory if it doesn't exist
		if err := os.MkdirAll("logs", 0755); err != nil {
			panic(fmt.Sprintf("failed to create logs directory: %v", err))
		}

		// Use current date for filename
		currentDate := time.Now().Format("2006-01-02")
		filename := fmt.Sprintf("logs/%s.log", currentDate)

		// Setup file rotation
		fileLogger := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    100, // megabytes
			MaxBackups: 30,  // keep 30 days of logs
			MaxAge:     30,  // days
			Compress:   true,
		}

		// Create multi-writer for both file and stdout
		multiWriter = io.MultiWriter(os.Stdout, fileLogger)
	}

	// Get log ID from context or generate new one
	logID := GetLogIDFromContext(ctx)
	if logID == "" {
		logID = GenerateLogID()
	}

	// Setup JSON handler with Grafana-friendly format and context-aware log ID
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: replacer,
	})).With(
		string(ServiceKey), serviceName,
		logIDString, logID,
	)

	slog.SetDefault(logger)
}

// NewLogger creates and returns a new logger instance
func NewLogger() (*slog.Logger, error) {
	cfg := config.GetConfig()

	// Setup the logger
	SetupLogger(cfg.ServiceName, string(cfg.Mode))

	// Return the default logger
	return slog.Default(), nil
}
