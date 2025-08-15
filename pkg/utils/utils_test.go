package utils_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"go-clean-ddd-es-template/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	uuid1 := utils.GenerateUUID()
	uuid2 := utils.GenerateUUID()

	assert.NotEmpty(t, uuid1)
	assert.NotEmpty(t, uuid2)
	assert.NotEqual(t, uuid1, uuid2)
	assert.Len(t, uuid1, 36) // UUID v4 format
	assert.Len(t, uuid2, 36)
}

func TestGenerateRandomString(t *testing.T) {
	str1, err1 := utils.GenerateRandomString(10)
	str2, err2 := utils.GenerateRandomString(10)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, str1)
	assert.NotEmpty(t, str2)
	assert.NotEqual(t, str1, str2)
	assert.Len(t, str1, 10)
	assert.Len(t, str2, 10)
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "test@sub.example.com", true},
		{"valid email with plus", "test+tag@example.com", true},
		{"invalid email - no @", "testexample.com", false},
		{"invalid email - no domain", "test@", false},
		{"invalid email - no local part", "@example.com", false},
		{"invalid email - empty", "", false},
		{"invalid email - spaces", "test @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"invalid uuid - wrong format", "550e8400-e29b-41d4-a716-44665544000", false},
		{"invalid uuid - wrong characters", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"invalid uuid - empty", "", false},
		{"invalid uuid - random string", "not-a-uuid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidUUID(tt.uuid)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	now := time.Now()
	formatted := utils.FormatTimestamp(now)

	assert.NotEmpty(t, formatted)
	assert.Contains(t, formatted, "T")
	// RFC3339 format may not always contain "Z" depending on timezone
	assert.True(t, strings.Contains(formatted, "Z") || strings.Contains(formatted, "+") || strings.Contains(formatted, "-"))
}

func TestParseTimestamp(t *testing.T) {
	expected := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	timestampStr := "2023-01-01T12:00:00Z"

	parsed, err := utils.ParseTimestamp(timestampStr)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)
}

func TestToJSON(t *testing.T) {
	data := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	jsonStr, err := utils.ToJSON(data)
	assert.NoError(t, err)
	assert.Contains(t, jsonStr, "John")
	assert.Contains(t, jsonStr, "30")
}

func TestFromJSON(t *testing.T) {
	jsonStr := `{"name":"John","age":30}`
	var data map[string]interface{}

	err := utils.FromJSON(jsonStr, &data)
	assert.NoError(t, err)
	assert.Equal(t, "John", data["name"])
	assert.Equal(t, float64(30), data["age"])
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "hello world", "hello world"},
		{"string with spaces", "  hello world  ", "hello world"},
		{"string with newlines", "hello\nworld", "hello\nworld"},
		{"string with tabs", "hello\tworld", "hello\tworld"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
		{"special characters", "hello!@#$%^&*()world", "hello!world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello world", 11, "hello world"},
		{"long string", "hello world", 5, "hello..."},
		{"empty string", "", 10, ""},
		{"zero max length", "hello", 0, "..."},
		{"negative max length", "hello", -1, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.TruncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"milliseconds", 500 * time.Millisecond, "500ms"},
		{"seconds", 2 * time.Second, "2.00s"},
		{"minutes", 2*time.Minute + 30*time.Second, "2m 30s"},
		{"hours", 2*time.Hour + 30*time.Minute, "2h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	tests := []struct {
		name     string
		item     string
		expected bool
	}{
		{"contains item", "banana", true},
		{"does not contain item", "orange", false},
		{"case sensitive", "Apple", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Contains(slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "with duplicates",
			input:    []string{"apple", "banana", "apple", "cherry", "banana"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.RemoveDuplicates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetry(t *testing.T) {
	tests := []struct {
		name             string
		maxAttempts      int
		delay            time.Duration
		failCount        int
		expectError      bool
		expectedAttempts int
	}{
		{
			name:             "success on first attempt",
			maxAttempts:      3,
			delay:            10 * time.Millisecond,
			failCount:        0,
			expectError:      false,
			expectedAttempts: 1,
		},
		{
			name:             "success on second attempt",
			maxAttempts:      3,
			delay:            10 * time.Millisecond,
			failCount:        1,
			expectError:      false,
			expectedAttempts: 2,
		},
		{
			name:             "success on last attempt",
			maxAttempts:      3,
			delay:            10 * time.Millisecond,
			failCount:        2,
			expectError:      false,
			expectedAttempts: 3,
		},
		{
			name:             "failure after all attempts",
			maxAttempts:      3,
			delay:            10 * time.Millisecond,
			failCount:        5,
			expectError:      true,
			expectedAttempts: 3,
		},
		{
			name:             "single attempt success",
			maxAttempts:      1,
			delay:            10 * time.Millisecond,
			failCount:        0,
			expectError:      false,
			expectedAttempts: 1,
		},
		{
			name:             "single attempt failure",
			maxAttempts:      1,
			delay:            10 * time.Millisecond,
			failCount:        1,
			expectError:      true,
			expectedAttempts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attemptCount := 0
			fn := func() error {
				attemptCount++
				if attemptCount <= tt.failCount {
					return fmt.Errorf("attempt %d failed", attemptCount)
				}
				return nil
			}

			err := utils.Retry(tt.maxAttempts, tt.delay, fn)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), fmt.Sprintf("failed after %d attempts", tt.maxAttempts))
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedAttempts, attemptCount)
		})
	}
}

func TestRetry_ZeroDelay(t *testing.T) {
	attemptCount := 0
	fn := func() error {
		attemptCount++
		if attemptCount < 3 {
			return fmt.Errorf("attempt %d failed", attemptCount)
		}
		return nil
	}

	start := time.Now()
	err := utils.Retry(3, 0, fn)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 3, attemptCount)
	// With zero delay, execution should be very fast
	assert.Less(t, duration, 100*time.Millisecond)
}

func TestRetry_WithDelay(t *testing.T) {
	attemptCount := 0
	fn := func() error {
		attemptCount++
		if attemptCount < 2 {
			return fmt.Errorf("attempt %d failed", attemptCount)
		}
		return nil
	}

	delay := 50 * time.Millisecond
	start := time.Now()
	err := utils.Retry(3, delay, fn)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 2, attemptCount)
	// Should have at least one delay
	assert.GreaterOrEqual(t, duration, delay)
	// But not more than 2 delays (since we succeed on attempt 2)
	assert.Less(t, duration, 3*delay)
}
