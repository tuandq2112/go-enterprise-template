package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode represents a unique error code
type ErrorCode string

// Common error codes
const (
	// Domain errors
	ErrInvalidEmail      ErrorCode = "INVALID_EMAIL"
	ErrInvalidName       ErrorCode = "INVALID_NAME"
	ErrInvalidUserID     ErrorCode = "INVALID_USER_ID"
	ErrUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrUserAlreadyExists ErrorCode = "USER_ALREADY_EXISTS"
	ErrUserDeleted       ErrorCode = "USER_DELETED"

	// Application errors
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCommandFailed    ErrorCode = "COMMAND_FAILED"
	ErrQueryFailed      ErrorCode = "QUERY_FAILED"

	// Infrastructure errors
	ErrDatabaseConnection  ErrorCode = "DATABASE_CONNECTION"
	ErrDatabaseQuery       ErrorCode = "DATABASE_QUERY"
	ErrDatabaseTransaction ErrorCode = "DATABASE_TRANSACTION"
	ErrEventStoreFailed    ErrorCode = "EVENT_STORE_FAILED"
	ErrEventPublishFailed  ErrorCode = "EVENT_PUBLISH_FAILED"
	ErrMessageBrokerFailed ErrorCode = "MESSAGE_BROKER_FAILED"

	// System errors
	ErrInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrTimeout            ErrorCode = "TIMEOUT"
	ErrUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrForbidden          ErrorCode = "FORBIDDEN"
	ErrNotFound           ErrorCode = "NOT_FOUND"
	ErrBadRequest         ErrorCode = "BAD_REQUEST"
)

// AppError represents an application error with i18n support
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	HTTPStatus int                    `json:"-"`
	Locale     string                 `json:"-"`
	Stack      []string               `json:"stack,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithCause sets the underlying cause of the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithLocale sets the locale for i18n
func (e *AppError) WithLocale(locale string) *AppError {
	e.Locale = locale
	return e
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithStack adds stack trace to the error
func (e *AppError) WithStack() *AppError {
	e.Stack = getStackTrace()
	return e
}

// New creates a new AppError with stack trace
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		Stack:      getStackTrace(),
	}
}

// Newf creates a new AppError with formatted message and stack trace
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		HTTPStatus: getHTTPStatus(code),
		Stack:      getStackTrace(),
	}
}

// Wrap wraps an existing error with additional context and stack trace
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      err,
		HTTPStatus: getHTTPStatus(code),
		Stack:      getStackTrace(),
	}
}

// Wrapf wraps an existing error with formatted message and stack trace
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Cause:      err,
		HTTPStatus: getHTTPStatus(code),
		Stack:      getStackTrace(),
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError converts an error to AppError if possible
func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}

	// If it's already an AppError, return it
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}

	// Try to parse error message as AppError (for backward compatibility)
	if strings.Contains(err.Error(), ": ") {
		parts := strings.SplitN(err.Error(), ": ", 2)
		if len(parts) == 2 {
			appErr := &AppError{
				Code:    ErrorCode(parts[0]),
				Message: parts[1],
				Cause:   err,
			}
			return appErr, true
		}
	}

	// Create a generic AppError wrapper
	appErr := &AppError{
		Code:    ErrInternalServer,
		Message: err.Error(),
		Cause:   err,
		Stack:   getStackTrace(),
	}
	return appErr, true
}

// getStackTrace returns the current stack trace
func getStackTrace() []string {
	var stack []string
	for i := 1; i < 10; i++ { // Skip first few frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	return stack
}

// getHTTPStatus returns the appropriate HTTP status code for an error code
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrBadRequest, ErrInvalidEmail, ErrInvalidName, ErrInvalidUserID, ErrValidationFailed:
		return 400
	case ErrUnauthorized:
		return 401
	case ErrForbidden:
		return 403
	case ErrNotFound, ErrUserNotFound:
		return 404
	case ErrUserAlreadyExists:
		return 409
	case ErrUserDeleted:
		return 410
	case ErrTimeout:
		return 408
	case ErrServiceUnavailable:
		return 503
	case ErrInternalServer, ErrDatabaseConnection, ErrDatabaseQuery, ErrDatabaseTransaction,
		ErrEventStoreFailed, ErrEventPublishFailed, ErrMessageBrokerFailed, ErrCommandFailed, ErrQueryFailed:
		return 500
	default:
		return 500
	}
}

// Common error constructors
func InvalidEmail(email string) *AppError {
	return New(ErrInvalidEmail, fmt.Sprintf("Invalid email format: %s", email))
}

func InvalidName(name string) *AppError {
	return New(ErrInvalidName, fmt.Sprintf("Invalid name: %s", name))
}

func InvalidUserID(userID string) *AppError {
	return New(ErrInvalidUserID, fmt.Sprintf("Invalid user ID: %s", userID))
}

func UserNotFound(userID string) *AppError {
	return New(ErrUserNotFound, fmt.Sprintf("User not found: %s", userID))
}

func UserAlreadyExists(email string) *AppError {
	return New(ErrUserAlreadyExists, fmt.Sprintf("User already exists with email: %s", email))
}

func UserDeleted(userID string) *AppError {
	return New(ErrUserDeleted, fmt.Sprintf("User is deleted: %s", userID))
}

func ValidationFailed(field string, reason string) *AppError {
	return New(ErrValidationFailed, fmt.Sprintf("Validation failed for %s: %s", field, reason))
}

func DatabaseError(operation string, err error) *AppError {
	return Wrap(err, ErrDatabaseQuery, fmt.Sprintf("Database %s failed", operation))
}

func EventStoreError(operation string, err error) *AppError {
	return Wrap(err, ErrEventStoreFailed, fmt.Sprintf("Event store %s failed", operation))
}

func EventPublishError(err error) *AppError {
	return Wrap(err, ErrEventPublishFailed, "Failed to publish event")
}

func MessageBrokerError(operation string, err error) *AppError {
	return Wrap(err, ErrMessageBrokerFailed, fmt.Sprintf("Message broker %s failed", operation))
}
