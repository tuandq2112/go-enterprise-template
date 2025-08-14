package validation

import (
	"errors"
	"regexp"
	"strings"

	"go-clean-ddd-es-template/pkg/utils"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidationResult represents validation result
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors"`
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
	}
}

// AddError adds an error to validation result
func (r *ValidationResult) AddError(field, message string) {
	r.IsValid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	if !utils.IsValidEmail(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// ValidateName validates name
func ValidateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if len(name) < 2 {
		return errors.New("name must be at least 2 characters")
	}

	if len(name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	// Check for valid characters
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(name) {
		return errors.New("name contains invalid characters")
	}

	return nil
}

// ValidateUUID validates UUID format
func ValidateUUID(uuidStr string) error {
	if uuidStr == "" {
		return errors.New("UUID is required")
	}

	if !utils.IsValidUUID(uuidStr) {
		return errors.New("invalid UUID format")
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return errors.New("password must be less than 128 characters")
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one digit")
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// ValidateStringLength validates string length
func ValidateStringLength(value, fieldName string, min, max int) error {
	if value == "" && min > 0 {
		return errors.New(fieldName + " is required")
	}

	if len(value) < min {
		return errors.New(fieldName + " must be at least " + string(rune(min)) + " characters")
	}

	if len(value) > max {
		return errors.New(fieldName + " must be less than " + string(rune(max)) + " characters")
	}

	return nil
}

// ValidateURL validates URL format
func ValidateURL(url string) error {
	if url == "" {
		return errors.New("URL is required")
	}

	urlRegex := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
	if !urlRegex.MatchString(url) {
		return errors.New("invalid URL format")
	}

	return nil
}

// ValidatePhoneNumber validates phone number format
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return errors.New("phone number is required")
	}

	// Remove spaces and special characters
	phone = regexp.MustCompile(`[\s\-\(\)]`).ReplaceAllString(phone, "")

	// Check if it's a valid phone number (simplified)
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone number format")
	}

	return nil
}

// ValidateDate validates date format
func ValidateDate(dateStr string) error {
	if dateStr == "" {
		return errors.New("date is required")
	}

	// Try to parse timestamp
	if _, err := utils.ParseTimestamp(dateStr); err == nil {
		return nil
	}

	return errors.New("invalid date format")
}

// SanitizeInput sanitizes user input
func SanitizeInput(input string) string {
	// Remove HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	input = htmlRegex.ReplaceAllString(input, "")

	// Remove script tags
	scriptRegex := regexp.MustCompile(`<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}
