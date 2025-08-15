package auth

import "errors"

// Auth-related errors
var (
	ErrPasswordTooShort        = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong         = errors.New("password must be no more than 128 characters long")
	ErrPasswordNoUpperCase     = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowerCase     = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit         = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecialChar   = errors.New("password must contain at least one special character")
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserInactive            = errors.New("user account is inactive")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token has expired")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)
