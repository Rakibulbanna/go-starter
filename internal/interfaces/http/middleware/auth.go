package middleware

import (
	"context"
	"fmt"
	"net/http"
	"skoolz/config"
	AppError "skoolz/internal/shared/error"
	"skoolz/internal/shared/types"
	"skoolz/internal/shared/utils"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// UserContext represents user information in request context
type UserContext struct {
	UserID   string
	Email    string
	Name     string
	Role     string
	IsActive bool
}

// context key for user information
type ctxKeyUser struct{}

// AuthMiddleware validates JWT tokens and extracts user information
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenString, err := extractTokenFromHeader(r)
		if err != nil {
			AppError.Unauthorized("Unauthorized: Invalid token format").WriteToResponse(w)
			return
		}

		// Validate and parse the token
		claims, err := validateToken(tokenString)
		if err != nil {
			AppError.Unauthorized("Unauthorized: Invalid token").WriteToResponse(w)
			return
		}

		// Check if user is active
		if !claims.IsActive {
			AppError.Unauthorized("Unauthorized: User is inactive").WriteToResponse(w)
			return
		}

		// Add user information to request context
		userCtx := UserContext{
			UserID:   claims.UserID,
			Email:    claims.Email,
			Name:     claims.Name,
			Role:     claims.Role,
			IsActive: claims.IsActive,
		}

		ctx := context.WithValue(r.Context(), ctxKeyUser{}, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware validates JWT tokens if present, but doesn't require them
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token from Authorization header
		tokenString, err := extractTokenFromHeader(r)
		if err != nil {
			// No token provided, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Validate and parse the token
		claims, err := validateToken(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Check if user is active
		if !claims.IsActive {
			// Inactive user, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Add user information to request context
		userCtx := UserContext{
			UserID:   claims.UserID,
			Email:    claims.Email,
			Name:     claims.Name,
			Role:     claims.Role,
			IsActive: claims.IsActive,
		}

		ctx := context.WithValue(r.Context(), ctxKeyUser{}, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware checks if the authenticated user has the required role
func RoleMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := utils.GetUserFromRequest(r)
			if !ok {
				AppError.Unauthorized("Unauthorized: User not authenticated").WriteToResponse(w)
				return
			}

			// for _, role := range requiredRoles {
			// 	if userCtx.Role == role {
			// 		hasRole = true
			// 		break
			// 	}
			// }
			// Check if user has any of the required roles
			hasRole := slices.Contains(requiredRoles, userCtx.Role)

			if !hasRole {
				AppError.Forbidden("Forbidden: Insufficient permissions").WriteToResponse(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractTokenFromHeader extracts the JWT token from the Authorization header
func extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", AppError.Unauthorized("Unauthorized: No token provided")
	}

	// Check if the header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", AppError.Unauthorized("Unauthorized: Invalid token format")
	}

	// Extract the token part (remove "Bearer " prefix)
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", AppError.Unauthorized("Unauthorized: Invalid token")
	}

	return token, nil
}

// validateToken validates and parses the JWT token
func validateToken(tokenString string) (*types.UserClaims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &types.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*types.UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, AppError.Unauthorized("Unauthorized: Invalid token")
}
