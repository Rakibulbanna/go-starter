# HTTP Middleware System

This package provides a comprehensive middleware system for HTTP requests with common middleware patterns and a flexible manager.

## Features

- **Middleware Manager**: Centralized middleware management with global and route-specific middleware
- **Common Middleware**: Pre-built middleware for common use cases
- **Structured Logging**: Request logging with structured data
- **Security**: Built-in security headers and CORS support
- **Request Tracking**: Unique request ID generation and tracking
- **Error Handling**: Panic recovery and error logging

## Quick Start

```go
// Create a new middleware manager (includes common middleware by default)
manager := middleware.NewManager()

// Apply to a handler
handler := manager.With(http.HandlerFunc(yourHandler))

// Or add custom middleware for specific routes
handler := manager.With(
    http.HandlerFunc(yourHandler),
    middleware.AuthMiddleware,
    middleware.RateLimitMiddleware(100),
)
```

## Available Middleware

### Default Middleware (Applied Automatically)

1. **RecoveryMiddleware**: Recovers from panics and logs errors
2. **RequestIDMiddleware**: Generates unique request IDs
3. **LoggingMiddleware**: Structured request logging
4. **SecurityHeadersMiddleware**: Adds security headers
5. **CORSMiddleware**: Handles CORS requests
6. **TimeoutMiddleware**: Adds request timeout (30 seconds)

### Additional Middleware

- **AuthMiddleware**: Basic authentication example
- **RateLimitMiddleware**: Simple rate limiting
- **ContentTypeMiddleware**: Content type validation

## Usage Examples

### Basic Usage

```go
// The manager automatically includes common middleware
manager := middleware.NewManager()

// All routes will have common middleware applied
mux.Handle("GET /", manager.With(http.HandlerFunc(handler)))
```

### Custom Middleware for Specific Routes

```go
manager := middleware.NewManager()

// Public route with only default middleware
mux.Handle("GET /public", manager.With(http.HandlerFunc(publicHandler)))

// Protected route with authentication
mux.Handle("GET /protected", manager.With(
    http.HandlerFunc(protectedHandler),
    middleware.AuthMiddleware,
))

// API route with rate limiting
mux.Handle("POST /api/data", manager.With(
    http.HandlerFunc(apiHandler),
    middleware.AuthMiddleware,
    middleware.RateLimitMiddleware(60),
    middleware.ContentTypeMiddleware("application/json"),
))
```

### Custom Middleware Order

```go
manager := middleware.NewManager()

// Clear default middleware and add custom order
manager.Clear().Use(
    middleware.RecoveryMiddleware,
    middleware.RequestIDMiddleware,
    middleware.LoggingMiddleware,
    // Add your custom middleware here
)
```

## Middleware Order

Middleware is applied in the following order:

1. Route-specific middleware (first to last)
2. Global middleware (first to last)

This means global middleware wraps route-specific middleware.

## Logging

The logging middleware provides structured logs with the following fields:

- `method`: HTTP method
- `path`: Request path
- `request_id`: Unique request identifier
- `ip`: Client IP address
- `user_agent`: User agent string
- `status`: Response status code
- `latency`: Request duration in seconds

## Security Headers

The security middleware adds the following headers:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

## CORS Configuration

The CORS middleware is configured for development with:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With`

For production, customize the CORS settings in the middleware.

## Creating Custom Middleware

```go
func CustomMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing logic here

        next.ServeHTTP(w, r)

        // Post-processing logic here
    })
}

// Usage
manager.Use(CustomMiddleware)
```

## Best Practices

1. **Order Matters**: Place critical middleware (auth, rate limiting) early in the chain
2. **Error Handling**: Always use RecoveryMiddleware for panic recovery
3. **Logging**: Use structured logging for better observability
4. **Security**: Apply security headers to all routes
5. **Performance**: Keep middleware lightweight and avoid blocking operations
6. **Testing**: Test middleware in isolation and as part of the request chain
