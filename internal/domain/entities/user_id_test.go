package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserID(t *testing.T) {
	userID := NewUserID()

	assert.NotEmpty(t, userID.value)
	assert.Len(t, userID.value, 36) // UUID length
	assert.False(t, userID.IsZero())
}

func TestNewUserIDFromString(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid UUID",
			id:      "123e4567-e89b-12d3-a456-426614174000",
			wantErr: false,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "invalid UUID format",
			id:      "invalid-uuid",
			wantErr: true,
		},
		{
			name:    "random string",
			id:      "random-string-123",
			wantErr: true,
		},
		{
			name:    "UUID with wrong length",
			id:      "123e4567-e89b-12d3-a456",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := NewUserIDFromString(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, userID.IsZero())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.id, userID.String())
			assert.Equal(t, tt.id, userID.Value())
			assert.False(t, userID.IsZero())
		})
	}
}

func TestUserID_String(t *testing.T) {
	expectedID := "123e4567-e89b-12d3-a456-426614174000"
	userID, err := NewUserIDFromString(expectedID)
	assert.NoError(t, err)

	assert.Equal(t, expectedID, userID.String())
}

func TestUserID_Value(t *testing.T) {
	expectedID := "123e4567-e89b-12d3-a456-426614174000"
	userID, err := NewUserIDFromString(expectedID)
	assert.NoError(t, err)

	assert.Equal(t, expectedID, userID.Value())
}

func TestUserID_Equals(t *testing.T) {
	id1, err := NewUserIDFromString("123e4567-e89b-12d3-a456-426614174000")
	assert.NoError(t, err)

	id2, err := NewUserIDFromString("123e4567-e89b-12d3-a456-426614174000")
	assert.NoError(t, err)

	id3, err := NewUserIDFromString("987fcdeb-51a2-43d1-b789-987654321000")
	assert.NoError(t, err)

	// Test equality with same ID
	assert.True(t, id1.Equals(id1))
	assert.True(t, id1.Equals(id2))

	// Test equality with different ID
	assert.False(t, id1.Equals(id3))

	// Test equality with zero value
	zeroID := UserID{}
	assert.False(t, id1.Equals(zeroID))
	assert.False(t, zeroID.Equals(id1))
}

func TestUserID_IsZero(t *testing.T) {
	// Test zero value
	zeroID := UserID{}
	assert.True(t, zeroID.IsZero())

	// Test non-zero value
	userID := NewUserID()
	assert.False(t, userID.IsZero())

	// Test valid UUID
	validID, err := NewUserIDFromString("123e4567-e89b-12d3-a456-426614174000")
	assert.NoError(t, err)
	assert.False(t, validID.IsZero())
}

func TestMustNewUserIDFromString(t *testing.T) {
	// Test valid UUID - should not panic
	validID := "123e4567-e89b-12d3-a456-426614174000"
	userID := MustNewUserIDFromString(validID)
	assert.Equal(t, validID, userID.String())

	// Test invalid UUID - should panic
	assert.Panics(t, func() {
		MustNewUserIDFromString("invalid-uuid")
	})
}

func TestUserID_Uniqueness(t *testing.T) {
	// Test that multiple calls to NewUserID generate different IDs
	ids := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		userID := NewUserID()
		idStr := userID.String()

		// Check for uniqueness
		assert.False(t, ids[idStr], "Duplicate ID generated: %s", idStr)
		ids[idStr] = true

		// Validate UUID format
		assert.Len(t, idStr, 36)
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, idStr)
	}
}

func TestUserID_Consistency(t *testing.T) {
	// Test that String() and Value() return the same result
	userID := NewUserID()

	assert.Equal(t, userID.String(), userID.Value())
	assert.Equal(t, userID.Value(), userID.String())
}
