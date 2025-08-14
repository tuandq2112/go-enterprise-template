package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
		expectedError bool
	}{
		{
			name:          "valid name",
			input:         "John Doe",
			expected:      "John Doe",
			expectedError: false,
		},
		{
			name:          "valid name with hyphen",
			input:         "Jean-Pierre",
			expected:      "Jean-Pierre",
			expectedError: false,
		},
		{
			name:          "valid name with apostrophe",
			input:         "O'Connor",
			expected:      "O'Connor",
			expectedError: false,
		},
		{
			name:          "name with leading spaces",
			input:         "  John Doe",
			expected:      "John Doe",
			expectedError: false,
		},
		{
			name:          "name with trailing spaces",
			input:         "John Doe  ",
			expected:      "John Doe",
			expectedError: false,
		},
		{
			name:          "empty name",
			input:         "",
			expected:      "",
			expectedError: true,
		},
		{
			name:          "whitespace only",
			input:         "   ",
			expected:      "",
			expectedError: true,
		},
		{
			name:          "single character",
			input:         "J",
			expected:      "",
			expectedError: true,
		},
		{
			name:          "name too long",
			input:         string(make([]byte, 101)),
			expected:      "",
			expectedError: true,
		},
		{
			name:          "name with numbers",
			input:         "John123",
			expected:      "",
			expectedError: true,
		},
		{
			name:          "name with special characters",
			input:         "John@Doe",
			expected:      "",
			expectedError: true,
		},
		{
			name:          "name with consecutive spaces",
			input:         "John  Doe",
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := NewName(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, Name{}, name)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, name.String())
				assert.Equal(t, tt.expected, name.Value())
			}
		})
	}
}

func TestName_Equals(t *testing.T) {
	name1, _ := NewName("John Doe")
	name2, _ := NewName("John Doe")
	name3, _ := NewName("Jane Doe")

	assert.True(t, name1.Equals(name2))
	assert.False(t, name1.Equals(name3))
}

func TestMustNewName(t *testing.T) {
	// Should not panic for valid name
	assert.NotPanics(t, func() {
		name := MustNewName("John Doe")
		assert.Equal(t, "John Doe", name.String())
	})

	// Should panic for invalid name
	assert.Panics(t, func() {
		MustNewName("")
	})
}

func TestName_String(t *testing.T) {
	name, _ := NewName("John Doe")
	assert.Equal(t, "John Doe", name.String())
}

func TestName_Value(t *testing.T) {
	name, _ := NewName("John Doe")
	assert.Equal(t, "John Doe", name.Value())
}
