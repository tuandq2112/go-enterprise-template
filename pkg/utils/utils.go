package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidUUID validates UUID format
func IsValidUUID(uuidStr string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuidStr))
}

// FormatTimestamp formats timestamp to RFC3339
func FormatTimestamp(timestamp time.Time) string {
	return timestamp.Format(time.RFC3339)
}

// ParseTimestamp parses RFC3339 timestamp
func ParseTimestamp(timestampStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timestampStr)
}

// ToJSON converts struct to JSON string
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON converts JSON string to struct
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// TruncateString truncates string to specified length
func TruncateString(s string, maxLength int) string {
	if maxLength <= 0 {
		return "..."
	}
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// SanitizeString removes special characters and trims whitespace
func SanitizeString(s string) string {
	// Remove special characters except alphanumeric, spaces, and common punctuation
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s.,!?-]`)
	sanitized := reg.ReplaceAllString(s, "")
	return strings.TrimSpace(sanitized)
}

// FormatDuration formats duration to human readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// Retry executes function with retry logic
func Retry(maxAttempts int, delay time.Duration, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxAttempts {
				time.Sleep(delay)
			}
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// Contains checks if slice contains element
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from slice
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
