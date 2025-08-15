package entities

import (
	"strings"
	"unicode"

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
	return Name{value: normalizeName(name)}, nil
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

// validateName validates name format with enhanced security
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

	// Check for valid characters (Unicode letters, spaces, hyphens, apostrophes, dots)
	if !isValidNameString(trimmedName) {
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

	// Check for suspicious patterns (security)
	if containsSuspiciousNamePatterns(trimmedName) {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_SUSPICIOUS_PATTERN", "en"))
	}

	// Check for control characters
	if containsControlCharacters(trimmedName) {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_CONTROL_CHARS", "en"))
	}

	// Check for excessive punctuation
	if hasExcessivePunctuation(trimmedName) {
		return errors.New(errors.ErrInvalidName, i18n.T("NAME_EXCESSIVE_PUNCTUATION", "en"))
	}

	return nil
}

// normalizeName normalizes the name (trim, normalize spaces, etc.)
func normalizeName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Normalize spaces (replace multiple spaces with single space)
	spaceRegex := strings.NewReplacer("  ", " ")
	for strings.Contains(name, "  ") {
		name = spaceRegex.Replace(name)
	}

	return name
}

// isValidNameString checks if the name contains only valid characters
func isValidNameString(name string) bool {
	for _, char := range name {
		if !isValidNameChar(char) {
			return false
		}
	}
	return true
}

// isValidNameChar checks if character is valid in a name
func isValidNameChar(char rune) bool {
	// Allow Unicode letters (including accented characters)
	if unicode.IsLetter(char) {
		return true
	}

	// Allow digits (for names like "John2" or "Mary3")
	if unicode.IsDigit(char) {
		return true
	}

	// Allow common name separators
	switch char {
	case ' ', '-', '\'', '.', ',', '(', ')':
		return true
	}

	return false
}

// containsSuspiciousNamePatterns checks for potentially malicious patterns in names
func containsSuspiciousNamePatterns(name string) bool {
	suspiciousPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"<iframe", "<object", "<embed", "data:text/html",
		"../../", "..\\", "file://", "ftp://", "gopher://",
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"UNION", "OR", "AND", "WHERE", "FROM", "JOIN",
	}

	nameLower := strings.ToLower(name)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}

	return false
}

// containsControlCharacters checks for control characters
func containsControlCharacters(name string) bool {
	for _, char := range name {
		if unicode.IsControl(char) && char != '\t' && char != '\n' && char != '\r' {
			return true
		}
	}
	return false
}

// hasExcessivePunctuation checks for excessive punctuation
func hasExcessivePunctuation(name string) bool {
	punctuationCount := 0
	totalChars := 0

	for _, char := range name {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			totalChars++
		} else if char == '.' || char == ',' || char == '-' || char == '\'' {
			punctuationCount++
		}
	}

	// If more than 30% of characters are punctuation, it's excessive
	if totalChars > 0 && float64(punctuationCount)/float64(totalChars) > 0.3 {
		return true
	}

	return false
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
