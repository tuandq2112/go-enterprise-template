package entities

import (
	"regexp"

	"go-clean-ddd-es-template/pkg/errors"
	"go-clean-ddd-es-template/pkg/i18n"
	"go-clean-ddd-es-template/pkg/utils"
)

// Email represents an email address value object
type Email struct {
	value string
}

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	if err := validateEmail(email); err != nil {
		return Email{}, err
	}
	return Email{value: email}, nil
}

// String returns the email as a string
func (e Email) String() string {
	return e.value
}

// Value returns the underlying email value
func (e Email) Value() string {
	return e.value
}

// Equals checks if two emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// validateEmail validates email format
func validateEmail(email string) error {
	if email == "" {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_REQUIRED", "en"))
	}

	if !utils.IsValidEmail(email) {
		return errors.New(errors.ErrInvalidEmail, i18n.T("INVALID_EMAIL", "en", email))
	}

	// Additional domain-specific validation
	if len(email) > 254 {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_TOO_LONG", "en"))
	}

	// Check for valid characters
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_INVALID_CHARS", "en"))
	}

	return nil
}

// MustNewEmail creates a new Email value object and panics if validation fails
// Use only in tests or when you're certain the email is valid
func MustNewEmail(email string) Email {
	e, err := NewEmail(email)
	if err != nil {
		panic(err)
	}
	return e
}
