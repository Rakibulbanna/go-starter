package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"skoolz/config"
	"skoolz/internal/logger"
	AppError "skoolz/internal/shared/error"
)

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Get request ID from context or generate one
		requestID := getRequestID(r.Context())

		// Log request start
		slog.Info("HTTP request started",
			logger.Method(r.Method),
			logger.Path(r.URL.Path),
			logger.RequestID(requestID),
			logger.Ip(getClientIP(r)),
			logger.UserAgent(r.UserAgent()),
		)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion
		slog.Info("HTTP request completed",
			logger.Method(r.Method),
			logger.Path(r.URL.Path),
			logger.Status(wrapped.statusCode),
			logger.Latency(duration),
			logger.RequestID(requestID),
		)
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", strings.Join(config.GetConfig().Cors.Origin, ","))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists
		if r.Header.Get("X-Request-ID") == "" {
			requestID := generateRequestID()
			r.Header.Set("X-Request-ID", requestID)
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))

		// Add request ID to context
		// Use a typed context key to avoid collisions and improve type safety
		type ctxKeyRequestID struct{}
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				slog.Error("HTTP request panic recovered",
					logger.Method(r.Method),
					logger.Path(r.URL.Path),
					logger.Extra(fmt.Sprintf("panic: %v", err)),
					logger.RequestID(getRequestID(r.Context())),
					"stack_trace", string(debug.Stack()),
				)

				// Return 500 Internal Server Error
				AppError.InternalServerError("Internal Server Error").WriteToResponse(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security-related headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getRequestID retrieves request ID from context
func getRequestID(ctx context.Context) string {
	// Use a typed context key to avoid collisions and improve type safety
	type ctxKeyRequestID struct{}
	if id, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return id
	}
	return generateRequestID()
}

// getClientIP extracts the real client IP address
func getClientIP(r *http.Request) string {
	// Check for forwarded headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to remote address
	return r.RemoteAddr
}

// RateLimitMiddleware is a simple rate limiting example
// In production, use a proper rate limiting library like golang.org/x/time/rate
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter (not suitable for production with multiple instances)
	clients := make(map[string][]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()

			// Clean old requests
			if times, exists := clients[clientIP]; exists {
				var validTimes []time.Time
				for _, t := range times {
					if now.Sub(t) < time.Minute {
						validTimes = append(validTimes, t)
					}
				}
				clients[clientIP] = validTimes
			}

			// Check rate limit
			if len(clients[clientIP]) >= requestsPerMinute {
				AppError.RateLimitExceeded("Rate Limit Exceeded").WriteToResponse(w)
				return
			}

			// Add current request
			clients[clientIP] = append(clients[clientIP], now)

			// Add rate limit headers for successful requests
			remaining := requestsPerMinute - len(clients[clientIP])
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(time.Minute).Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware ensures requests have the correct content type
func ContentTypeMiddleware(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				if r.Header.Get("Content-Type") != contentType {
					AppError.UnsupportedMediaType("Unsupported Media Type").WriteToResponse(w)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
