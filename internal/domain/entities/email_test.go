package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		expectedError bool
	}{
		{
			name:          "valid email",
			email:         "test@example.com",
			expectedError: false,
		},
		{
			name:          "valid email with subdomain",
			email:         "test@sub.example.com",
			expectedError: false,
		},
		{
			name:          "valid email with plus",
			email:         "test+tag@example.com",
			expectedError: false,
		},
		{
			name:          "empty email",
			email:         "",
			expectedError: true,
		},
		{
			name:          "invalid email format",
			email:         "invalid-email",
			expectedError: true,
		},
		{
			name:          "email without domain",
			email:         "test@",
			expectedError: true,
		},
		{
			name:          "email without local part",
			email:         "@example.com",
			expectedError: true,
		},
		{
			name:          "email with invalid characters",
			email:         "test<script>@example.com",
			expectedError: true,
		},
		{
			name:          "email too long",
			email:         "verylongemailaddress" + string(make([]byte, 300)) + "@example.com",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.email)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, Email{}, email)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.email, email.String())
				assert.Equal(t, tt.email, email.Value())
			}
		})
	}
}

func TestEmail_Equals(t *testing.T) {
	email1, _ := NewEmail("test@example.com")
	email2, _ := NewEmail("test@example.com")
	email3, _ := NewEmail("different@example.com")

	assert.True(t, email1.Equals(email2))
	assert.False(t, email1.Equals(email3))
}

func TestMustNewEmail(t *testing.T) {
	// Should not panic for valid email
	assert.NotPanics(t, func() {
		email := MustNewEmail("test@example.com")
		assert.Equal(t, "test@example.com", email.String())
	})

	// Should panic for invalid email
	assert.Panics(t, func() {
		MustNewEmail("invalid-email")
	})
}

func TestEmail_String(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	assert.Equal(t, "test@example.com", email.String())
}

func TestEmail_Value(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	assert.Equal(t, "test@example.com", email.Value())
}
