package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@subdomain.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			email:   "user.name@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "invalid email format",
			email:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "email without domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "email without local part",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "email with spaces",
			email:   "user @example.com",
			wantErr: true,
		},
		{
			name:    "email too long",
			email:   "verylongemailaddresswithlotsofcharactersandnumbers123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890@example.com",
			wantErr: true,
		},
		{
			name:    "email with invalid characters",
			email:   "user<script>@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, Email{}, email)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.email, email.String())
			assert.Equal(t, tt.email, email.Value())
		})
	}
}

func TestEmail_String(t *testing.T) {
	email, err := NewEmail("test@example.com")
	assert.NoError(t, err)

	assert.Equal(t, "test@example.com", email.String())
}

func TestEmail_Value(t *testing.T) {
	email, err := NewEmail("test@example.com")
	assert.NoError(t, err)

	assert.Equal(t, "test@example.com", email.Value())
}

func TestEmail_Equals(t *testing.T) {
	email1, err := NewEmail("test@example.com")
	assert.NoError(t, err)

	email2, err := NewEmail("test@example.com")
	assert.NoError(t, err)

	email3, err := NewEmail("different@example.com")
	assert.NoError(t, err)

	// Test equality with same email
	assert.True(t, email1.Equals(email1))
	assert.True(t, email1.Equals(email2))

	// Test equality with different email
	assert.False(t, email1.Equals(email3))

	// Test equality with zero value
	zeroEmail := Email{}
	assert.False(t, email1.Equals(zeroEmail))
	assert.False(t, zeroEmail.Equals(email1))
}

func TestMustNewEmail(t *testing.T) {
	// Test valid email - should not panic
	email := MustNewEmail("test@example.com")
	assert.Equal(t, "test@example.com", email.String())

	// Test invalid email - should panic
	assert.Panics(t, func() {
		MustNewEmail("invalid-email")
	})
}

func TestEmail_ZeroValue(t *testing.T) {
	// Test zero value behavior
	zeroEmail := Email{}

	assert.Equal(t, "", zeroEmail.String())
	assert.Equal(t, "", zeroEmail.Value())
	assert.True(t, zeroEmail.Equals(Email{}))
}

func TestEmail_CaseInsensitive(t *testing.T) {
	// Test that email comparison is case-sensitive (as per RFC 5321)
	email1, err := NewEmail("Test@Example.com")
	assert.NoError(t, err)

	email2, err := NewEmail("test@example.com")
	assert.NoError(t, err)

	// These should be considered different emails
	assert.False(t, email1.Equals(email2))
}

func TestEmail_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "email with dots in local part",
			email:   "user.name@example.com",
			wantErr: false,
		},
		{
			name:    "email with plus in local part",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "email with underscore in local part",
			email:   "user_name@example.com",
			wantErr: false,
		},
		{
			name:    "email with dash in domain",
			email:   "user@example-domain.com",
			wantErr: false,
		},
		{
			name:    "email with multiple dots in domain",
			email:   "user@sub.domain.example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.email, email.String())
		})
	}
}
