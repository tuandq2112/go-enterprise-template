package entities

import (
	"time"
)

// User represents a user entity with value objects
type User struct {
	ID        UserID    `json:"id"`
	Email     Email     `json:"email"`
	Name      Name      `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUser creates a new User entity with validation
func NewUser(email, name string) (*User, error) {
	// Validate and create Email value object
	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	// Validate and create Name value object
	nameVO, err := NewName(name)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:        NewUserID(),
		Email:     emailVO,
		Name:      nameVO,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateName updates the user's name with validation
func (u *User) UpdateName(name string) error {
	nameVO, err := NewName(name)
	if err != nil {
		return err
	}
	u.Name = nameVO
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateEmail updates the user's email with validation
func (u *User) UpdateEmail(email string) error {
	emailVO, err := NewEmail(email)
	if err != nil {
		return err
	}
	u.Email = emailVO
	u.UpdatedAt = time.Now()
	return nil
}

// GetEmail returns the email as a string
func (u *User) GetEmail() string {
	return u.Email.String()
}

// GetName returns the name as a string
func (u *User) GetName() string {
	return u.Name.String()
}

// GetID returns the user ID as a string
func (u *User) GetID() string {
	return u.ID.String()
}

// Equals checks if two users are equal
func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}
	return u.ID.Equals(other.ID) &&
		u.Email.Equals(other.Email) &&
		u.Name.Equals(other.Name)
}

// IsValid checks if the user entity is valid
func (u *User) IsValid() bool {
	return !u.ID.IsZero() &&
		u.Email.Value() != "" &&
		u.Name.Value() != ""
}
