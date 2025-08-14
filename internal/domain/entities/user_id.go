package entities

import (
	"github.com/google/uuid"

	"go-clean-ddd-es-template/pkg/errors"
	"go-clean-ddd-es-template/pkg/i18n"
)

// UserID represents a user identifier value object
type UserID struct {
	value string
}

// NewUserID creates a new UserID value object
func NewUserID() UserID {
	return UserID{value: uuid.New().String()}
}

// NewUserIDFromString creates a UserID from a string with validation
func NewUserIDFromString(id string) (UserID, error) {
	if err := validateUserID(id); err != nil {
		return UserID{}, err
	}
	return UserID{value: id}, nil
}

// String returns the user ID as a string
func (u UserID) String() string {
	return u.value
}

// Value returns the underlying user ID value
func (u UserID) Value() string {
	return u.value
}

// Equals checks if two user IDs are equal
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}

// IsZero checks if the user ID is zero value
func (u UserID) IsZero() bool {
	return u.value == ""
}

// validateUserID validates user ID format
func validateUserID(id string) error {
	if id == "" {
		return errors.New(errors.ErrInvalidUserID, i18n.T("USER_ID_REQUIRED", "en"))
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return errors.New(errors.ErrInvalidUserID, i18n.T("USER_ID_INVALID_FORMAT", "en"))
	}

	return nil
}

// MustNewUserIDFromString creates a UserID from a string and panics if validation fails
// Use only in tests or when you're certain the ID is valid
func MustNewUserIDFromString(id string) UserID {
	u, err := NewUserIDFromString(id)
	if err != nil {
		panic(err)
	}
	return u
}
