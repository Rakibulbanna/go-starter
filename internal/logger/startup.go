package logger

import (
	"fmt"
	"log/slog"
)

// StartupLogger provides concise startup messages with icons
type StartupLogger struct {
	logger *slog.Logger
}

// NewStartupLogger creates a new startup logger
func NewStartupLogger(logger *slog.Logger) *StartupLogger {
	return &StartupLogger{
		logger: logger,
	}
}

// Success logs a successful startup step with a green checkmark
func (s *StartupLogger) Success(message string, args ...any) {
	icon := "✅"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}

// Info logs an informational startup step with a blue info icon
func (s *StartupLogger) Info(message string, args ...any) {
	icon := "ℹ️"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}

// Warning logs a warning startup step with a yellow warning icon
func (s *StartupLogger) Warning(message string, args ...any) {
	icon := "⚠️"
	if len(args) > 0 {
		s.logger.Warn(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Warn(fmt.Sprintf("%s %s", icon, message))
	}
}

// Error logs an error startup step with a red error icon
func (s *StartupLogger) Error(message string, args ...any) {
	icon := "❌"
	if len(args) > 0 {
		s.logger.Error(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Error(fmt.Sprintf("%s %s", icon, message))
	}
}

// Service logs a service startup step with a rocket icon
func (s *StartupLogger) Service(message string, args ...any) {
	icon := "🚀"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}

// Database logs a database startup step with a database icon
func (s *StartupLogger) Database(message string, args ...any) {
	icon := "🗄️"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}

// Cache logs a cache startup step with a lightning icon
func (s *StartupLogger) Cache(message string, args ...any) {
	icon := "⚡"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}

// Messaging logs a messaging startup step with a message icon
func (s *StartupLogger) Messaging(message string, args ...any) {
	icon := "📨"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}

// Stream logs a stream startup step with a stream icon
func (s *StartupLogger) Stream(message string, args ...any) {
	icon := "🌊"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}

// Subscription logs a subscription startup step with a subscription icon
func (s *StartupLogger) Subscription(message string, args ...any) {
	icon := "📡"
	if len(args) > 0 {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message), args...)
	} else {
		s.logger.Info(fmt.Sprintf("%s %s", icon, message))
	}
}
