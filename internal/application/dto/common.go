package dto

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// PaginationRequest represents common pagination parameters
type PaginationRequest struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

// PaginationResponse represents common pagination response
type PaginationResponse struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationError represents validation error details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// ValidateRequest validates a request struct using validator
func ValidateRequest(req interface{}) error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Error())
		}
		return errors.New("validation failed: " + strings.Join(validationErrors, "; "))
	}
	return nil
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message, code string) *ErrorResponse {
	return &ErrorResponse{
		Error: message,
		Code:  code,
	}
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(message string, details map[string]string) *ErrorResponse {
	return &ErrorResponse{
		Error:   message,
		Code:    "VALIDATION_ERROR",
		Details: details,
	}
}
