# Response Package Documentation

This package provides a comprehensive response handling system for Go HTTP applications, including structured response types, builders, and convenient HTTP writing functions.

## Table of Contents

- [Overview](#overview)
- [Response Types](#response-types)
- [Response Builder](#response-builder)
- [HTTP Writing Functions](#http-writing-functions)
- [Use Case Examples](#use-case-examples)
- [Best Practices](#best-practices)

## Overview

The response package offers a standardized way to handle API responses with consistent structure, proper error handling, and convenient HTTP writing utilities.

## Response Types

### BaseResponse

The foundation for all response types:

```go
type BaseResponse struct {
    Status    ResponseStatus `json:"status"`
    Message   string         `json:"message"`
    Timestamp time.Time      `json:"timestamp"`
    RequestID string         `json:"request_id,omitempty"`
}
```

### SuccessResponse

For successful operations with optional data:

```go
type SuccessResponse struct {
    BaseResponse
    Data interface{} `json:"data,omitempty"`
}
```

### ErrorResponse

For error responses with details:

```go
type ErrorResponse struct {
    BaseResponse
    Error   string                 `json:"error"`
    Code    string                 `json:"code,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### PaginatedResponse

For paginated data responses:

```go
type PaginatedResponse struct {
    BaseResponse
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}
```

## Response Builder

The `ResponseBuilder` provides a fluent interface for creating responses:

```go
builder := NewResponseBuilder()

// Create success response
successResp := builder.Success("User created successfully", userData)

// Create error response
errorResp := builder.Error("Invalid input", "VALIDATION_ERROR", validationDetails)

// Create paginated response
paginatedResp := builder.Paginated("Users retrieved", users, 1, 10, 100)
```

## HTTP Writing Functions

### Basic Writing Functions

- `WriteSuccess(w, statusCode, message?, data?)` - Write success response
- `WriteError(w, statusCode, message?, errorCode?)` - Write error response
- `WriteErrorWithDetails(w, statusCode, message?, errorCode?, details?)` - Write error with details
- `WritePaginated(w, statusCode, message?, data?, page?, limit?, total?)` - Write paginated response
- `WriteJSON(w, statusCode, data)` - Write any JSON response

### Common HTTP Status Functions

- `WriteOK(w, message?, data?)` - 200 OK
- `WriteCreated(w, message?, data?)` - 201 Created
- `WriteNoContent(w)` - 204 No Content
- `WriteBadRequest(w, message?, errorCode?)` - 400 Bad Request
- `WriteUnauthorized(w, message?, errorCode?)` - 401 Unauthorized
- `WriteForbidden(w, message?, errorCode?)` - 403 Forbidden
- `WriteNotFound(w, message?, errorCode?)` - 404 Not Found
- `WriteConflict(w, message?, errorCode?)` - 409 Conflict
- `WriteUnprocessableEntity(w, message?, errorCode?, details?)` - 422 Unprocessable Entity
- `WriteInternalServerError(w, message?, errorCode?)` - 500 Internal Server Error
- `WriteServiceUnavailable(w, message?, errorCode?)` - 503 Service Unavailable

## Use Case Examples

### 1. User Management API

#### Create User Success

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    // ... user creation logic ...

    user := &User{
        ID:    "123",
        Name:  "John Doe",
        Email: "john@example.com",
    }

    // Using helper function
    response.WriteCreated(w, "User created successfully", user)

    // Alternative using builder
    // builder := NewResponseBuilder()
    // resp := builder.Success("User created successfully", user)
    // response.WriteJSON(w, http.StatusCreated, resp)
}
```

**Response:**

```json
{
  "status": "success",
  "message": "User created successfully",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "id": "123",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

#### Get User Success

```go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")

    user, err := userService.GetByID(userID)
    if err != nil {
        response.WriteNotFound(w, "User not found", "USER_NOT_FOUND")
        return
    }

    response.WriteOK(w, "User retrieved successfully", user)
}
```

#### Get User Not Found

```go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")

    user, err := userService.GetByID(userID)
    if err != nil {
        response.WriteNotFound(w, "User not found", "USER_NOT_FOUND")
        return
    }

    response.WriteOK(w, "User retrieved successfully", user)
}
```

**Response:**

```json
{
  "status": "error",
  "message": "User not found",
  "timestamp": "2024-01-15T10:30:00Z",
  "error": "USER_NOT_FOUND"
}
```

### 2. List Users with Pagination

```go
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
    page := 1
    limit := 10

    if pageStr := r.URL.Query().Get("page"); pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil {
            page = p
        }
    }

    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil {
            limit = l
        }
    }

    users, total, err := userService.List(page, limit)
    if err != nil {
        response.WriteInternalServerError(w, "Failed to retrieve users", "DATABASE_ERROR")
        return
    }

    response.WritePaginated(w, http.StatusOK, "Users retrieved successfully", users, page, limit, total)
}
```

**Response:**

```json
{
  "status": "success",
  "message": "Users retrieved successfully",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": [
    {
      "id": "123",
      "name": "John Doe",
      "email": "john@example.com"
    },
    {
      "id": "124",
      "name": "Jane Smith",
      "email": "jane@example.com"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

### 3. Validation Error with Details

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        response.WriteBadRequest(w, "Invalid JSON format", "INVALID_JSON")
        return
    }

    // Validate user data
    validationErrors := validateUser(user)
    if len(validationErrors) > 0 {
        details := map[string]interface{}{
            "validation_errors": validationErrors,
            "field_count":       len(validationErrors),
        }

        response.WriteUnprocessableEntity(w, "Validation failed", "VALIDATION_ERROR", details)
        return
    }

    // ... create user logic ...
}
```

**Response:**

```json
{
  "status": "error",
  "message": "Validation failed",
  "timestamp": "2024-01-15T10:30:00Z",
  "error": "VALIDATION_ERROR",
  "details": {
    "validation_errors": ["Email is required", "Name must be at least 2 characters"],
    "field_count": 2
  }
}
```

### 4. Authentication Error

```go
func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")
    if token == "" {
        response.WriteUnauthorized(w, "Authentication required", "MISSING_TOKEN")
        return
    }

    user, err := authService.ValidateToken(token)
    if err != nil {
        response.WriteUnauthorized(w, "Invalid token", "INVALID_TOKEN")
        return
    }

    // ... protected logic ...
}
```

### 5. Using Response Builder with Request ID

```go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-ID")

    user, err := userService.GetByID(userID)
    if err != nil {
        builder := NewResponseBuilder()
        errorResp := builder.Error("User not found", "USER_NOT_FOUND", nil)

        if requestID != "" {
            builder.WithRequestID(errorResp, requestID)
        }

        response.WriteJSON(w, http.StatusNotFound, errorResp)
        return
    }

    builder := NewResponseBuilder()
    successResp := builder.Success("User retrieved successfully", user)

    if requestID != "" {
        builder.WithRequestID(successResp, requestID)
    }

    response.WriteJSON(w, http.StatusOK, successResp)
}
```

**Response with Request ID:**

```json
{
  "status": "success",
  "message": "User retrieved successfully",
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-12345-abcde",
  "data": {
    "id": "123",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

### 6. Service Unavailable

```go
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    if err := database.Ping(); err != nil {
        response.WriteServiceUnavailable(w, "Database connection failed", "DB_CONNECTION_ERROR")
        return
    }

    if err := redis.Ping(); err != nil {
        response.WriteServiceUnavailable(w, "Cache service unavailable", "CACHE_ERROR")
        return
    }

    response.WriteOK(w, "Service is healthy")
}
```

### 7. Conflict Error

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)

    // Check if user already exists
    existingUser, _ := userService.GetByEmail(user.Email)
    if existingUser != nil {
        response.WriteConflict(w, "User with this email already exists", "EMAIL_CONFLICT")
        return
    }

    // ... create user logic ...
}
```

## Best Practices

### 1. Consistent Error Codes

Use standardized error codes across your application:

- `VALIDATION_ERROR` - Input validation failures
- `NOT_FOUND` - Resource not found
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `CONFLICT` - Resource conflicts
- `INTERNAL_SERVER_ERROR` - Server errors

### 2. Meaningful Messages

Provide clear, user-friendly messages:

```go
// Good
response.WriteBadRequest(w, "Email address is required", "VALIDATION_ERROR")

// Avoid
response.WriteBadRequest(w, "Error", "ERROR")
```

### 3. Use Appropriate HTTP Status Codes

- `200 OK` - Successful GET requests
- `201 Created` - Successful POST requests
- `204 No Content` - Successful DELETE requests
- `400 Bad Request` - Client errors
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Authorization failed
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflicts
- `422 Unprocessable Entity` - Validation errors
- `500 Internal Server Error` - Server errors

### 4. Include Request IDs for Debugging

```go
func AddRequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        ctx := context.WithValue(r.Context(), "request_id", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 5. Handle Panics Gracefully

```go
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                response.WriteInternalServerError(w, "Internal server error", "PANIC_RECOVERED")
            }
        }()

        next.ServeHTTP(w, r)
    })
}
```

### 6. Structured Logging

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Create a custom response writer to capture status code
        wrappedWriter := &responseWriter{ResponseWriter: w}

        next.ServeHTTP(wrappedWriter, r)

        duration := time.Since(start)

        log.Printf(
            "Method: %s, Path: %s, Status: %d, Duration: %v",
            r.Method, r.URL.Path, wrappedWriter.statusCode, duration,
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

This comprehensive response package provides everything you need to build consistent, well-structured API responses in your Go applications.
