package types

import "github.com/golang-jwt/jwt/v5"

// UserClaims represents the JWT claims for user authentication
type UserClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
	jwt.RegisteredClaims
}
