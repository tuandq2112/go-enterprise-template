package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userName string
		wantErr  bool
	}{
		{
			name:     "valid user",
			email:    "test@example.com",
			userName: "John Doe",
			wantErr:  false,
		},
		{
			name:     "invalid email",
			email:    "invalid-email",
			userName: "John Doe",
			wantErr:  true,
		},
		{
			name:     "empty name",
			email:    "test@example.com",
			userName: "",
			wantErr:  true,
		},
		{
			name:     "empty email",
			email:    "",
			userName: "John Doe",
			wantErr:  true,
		},
		{
			name:     "email with spaces",
			email:    "test @example.com",
			userName: "John Doe",
			wantErr:  true,
		},
		{
			name:     "name too long",
			email:    "test@example.com",
			userName: "This is a very long name that exceeds the maximum allowed length for a user name in our system and should be rejected because it is way too long and contains many characters that make it exceed the limit",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.userName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.email, user.GetEmail())
			assert.Equal(t, tt.userName, user.GetName())
			assert.NotEmpty(t, user.GetID())
			assert.False(t, user.CreatedAt.IsZero())
			assert.False(t, user.UpdatedAt.IsZero())
			assert.True(t, user.IsValid())
		})
	}
}

func TestUser_UpdateName(t *testing.T) {
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	originalUpdatedAt := user.UpdatedAt

	// Test valid name update
	err = user.UpdateName("Jane Doe")
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", user.GetName())
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// Test invalid name update
	err = user.UpdateName("")
	assert.Error(t, err)
	assert.Equal(t, "Jane Doe", user.GetName()) // Name should remain unchanged
}

func TestUser_UpdateEmail(t *testing.T) {
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	originalUpdatedAt := user.UpdatedAt

	// Test valid email update
	err = user.UpdateEmail("new@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", user.GetEmail())
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// Test invalid email update
	err = user.UpdateEmail("invalid-email")
	assert.Error(t, err)
	assert.Equal(t, "new@example.com", user.GetEmail()) // Email should remain unchanged
}

func TestUser_GetEmail(t *testing.T) {
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	assert.Equal(t, "test@example.com", user.GetEmail())
}

func TestUser_GetName(t *testing.T) {
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	assert.Equal(t, "John Doe", user.GetName())
}

func TestUser_GetID(t *testing.T) {
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	id := user.GetID()
	assert.NotEmpty(t, id)
	assert.Len(t, id, 36) // UUID length
}

func TestUser_Equals(t *testing.T) {
	user1, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user1)

	user2, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user2)

	user3, err := NewUser("different@example.com", "Jane Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user3)

	// Test equality with same user
	assert.True(t, user1.Equals(user1))

	// Test equality with different users (should be false due to different IDs)
	assert.False(t, user1.Equals(user2))

	// Test equality with nil
	assert.False(t, user1.Equals(nil))

	// Test equality with completely different user
	assert.False(t, user1.Equals(user3))
}

func TestUser_SetPasswordHash(t *testing.T) {
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	originalUpdatedAt := user.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure time difference

	hash := "hashed_password_123"
	user.SetPasswordHash(hash)

	assert.Equal(t, hash, user.GetPasswordHash())
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
}

func TestUser_GetPasswordHash(t *testing.T) {
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Initially should be empty
	assert.Empty(t, user.GetPasswordHash())

	// Set password hash
	hash := "hashed_password_123"
	user.SetPasswordHash(hash)

	assert.Equal(t, hash, user.GetPasswordHash())
}

func TestUser_IsValid(t *testing.T) {
	// Test valid user
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.True(t, user.IsValid())

	// Test invalid user with empty email
	user.Email = Email{}
	assert.False(t, user.IsValid())

	// Test invalid user with empty name
	user, err = NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	user.Name = Name{}
	assert.False(t, user.IsValid())

	// Test invalid user with zero ID
	user, err = NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	user.ID = UserID{}
	assert.False(t, user.IsValid())
}
