package utils

import (
	"context"
	"net/http"
	"skoolz/internal/shared/types"
)

type ctxKeyUser struct{}

// GetUserFromContext retrieves user information from the request context
func GetUserFromContext(ctx context.Context) (*types.UserClaims, bool) {
	user, ok := ctx.Value(ctxKeyUser{}).(*types.UserClaims)
	return user, ok
}

// GetUserFromRequest retrieves user information from the request context
func GetUserFromRequest(r *http.Request) (*types.UserClaims, bool) {
	return GetUserFromContext(r.Context())
}
