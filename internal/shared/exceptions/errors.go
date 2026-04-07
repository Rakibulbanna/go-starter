package exceptions

import "errors"

// Domain errors
var (
	ErrInvalidUserName       = errors.New("invalid user name")
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrPasswordTooShort      = errors.New("password too short (minimum 8 characters)")
	ErrPasswordTooLong       = errors.New("password too long (maximum 128 characters)")
	ErrPasswordTooWeak       = errors.New("password too weak (must contain uppercase, lowercase, number, and special character)")
	ErrPasswordHashingFailed = errors.New("failed to hash password")
	ErrInvalidRole           = errors.New("invalid role")
	ErrInvalidUUID           = errors.New("invalid UUID")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserInactive          = errors.New("user is inactive")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrValidationFailed      = errors.New("validation failed")
	ErrDomainViolation       = errors.New("domain rule violation")
	ErrAggregateNotFound     = errors.New("aggregate not found")
	ErrConcurrencyConflict   = errors.New("concurrency conflict")
	ErrEventPublishFailed    = errors.New("failed to publish domain event")
	ErrRepositoryError       = errors.New("repository error")
	ErrInfrastructureError   = errors.New("infrastructure error")
	ErrExternalServiceError  = errors.New("external service error")

	// Dummy domain errors
	ErrInvalidDummyName   = errors.New("invalid dummy name")
	ErrInvalidDummyStatus = errors.New("invalid dummy status")
	ErrDummyNotFound      = errors.New("dummy not found")
	ErrDummyAlreadyExists = errors.New("dummy already exists")
)
