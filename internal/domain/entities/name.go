package entities

import (
	"regexp"
	"strings"

	"go-clean-ddd-es-template/pkg/errors"
	"go-clean-ddd-es-template/pkg/i18n"
)

// Name represents a person's name value object
type Name struct {
	value string
}

// NewName creates a new Name value object with validation
func NewName(name string) (Name, error) {
	if err := validateName(name); err != nil {
		return Name{}, err
	}
	return Name{value: strings.TrimSpace(name)}, nil
}

// String returns the name as a string
func (n Name) String() string {
	return n.value
}

// Value returns the underlying name value
func (n Name) Value() string {
	return n.value
}

// Equals checks if two names are equal
func (n Name) Equals(other Name) bool {
	return n.value == other.value
}

// validateName validates name format
func validateName(name string) error {
	if name == "" {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_REQUIRED", "en"))
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_REQUIRED", "en"))
	}

	if len(trimmedName) < 2 {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_TOO_SHORT", "en"))
	}

	if len(trimmedName) > 100 {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_TOO_LONG", "en"))
	}

	// Check for valid characters (letters, spaces, hyphens, apostrophes)
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(trimmedName) {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_INVALID_CHARS", "en"))
	}

	// Check for consecutive spaces
	if strings.Contains(trimmedName, "  ") {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_CONSECUTIVE_SPACES", "en"))
	}

	// Check for leading/trailing spaces
	if strings.HasPrefix(trimmedName, " ") || strings.HasSuffix(trimmedName, " ") {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_LEADING_TRAILING_SPACES", "en"))
	}

	return nil
}

// MustNewName creates a new Name value object and panics if validation fails
// Use only in tests or when you're certain the name is valid
func MustNewName(name string) Name {
	n, err := NewName(name)
	if err != nil {
		panic(err)
	}
	return n
}
