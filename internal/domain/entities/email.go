package entities

import (
	"net/mail"
	"strings"
	"unicode"

	"go-clean-ddd-es-template/pkg/errors"
	"go-clean-ddd-es-template/pkg/i18n"
)

// Email represents an email address value object
type Email struct {
	value string
}

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	if err := validateEmail(email); err != nil {
		return Email{}, err
	}
	return Email{value: strings.ToLower(strings.TrimSpace(email))}, nil
}

// String returns the email as a string
func (e Email) String() string {
	return e.value
}

// Value returns the underlying email value
func (e Email) Value() string {
	return e.value
}

// Equals checks if two emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// validateEmail validates email format with enhanced security
func validateEmail(email string) error {
	if email == "" {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_REQUIRED", "en"))
	}

	// Trim whitespace
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_REQUIRED", "en"))
	}

	// Check length limits (RFC 5321)
	if len(email) > 254 {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_TOO_LONG", "en"))
	}

	// Check for minimum length
	if len(email) < 5 { // a@b.c
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_TOO_SHORT", "en"))
	}

	// Check for basic format
	if !strings.Contains(email, "@") {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_MISSING_AT", "en"))
	}

	// Split email into local and domain parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_INVALID_FORMAT", "en"))
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Validate local part
	if err := validateLocalPart(localPart); err != nil {
		return err
	}

	// Validate domain part
	if err := validateDomainPart(domainPart); err != nil {
		return err
	}

	// Use Go's built-in email validation as additional check
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_INVALID_FORMAT", "en"))
	}

	// Check for suspicious patterns (security)
	if containsSuspiciousPatterns(email) {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_SUSPICIOUS_PATTERN", "en"))
	}

	return nil
}

// validateLocalPart validates the local part of email
func validateLocalPart(localPart string) error {
	if localPart == "" {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_LOCAL_PART_EMPTY", "en"))
	}

	if len(localPart) > 64 {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_LOCAL_PART_TOO_LONG", "en"))
	}

	// Check for valid characters in local part
	for _, char := range localPart {
		if !isValidLocalPartChar(char) {
			return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_INVALID_CHARS", "en"))
		}
	}

	// Check for consecutive dots
	if strings.Contains(localPart, "..") {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_CONSECUTIVE_DOTS", "en"))
	}

	// Check for leading/trailing dots
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_LEADING_TRAILING_DOTS", "en"))
	}

	return nil
}

// validateDomainPart validates the domain part of email
func validateDomainPart(domainPart string) error {
	if domainPart == "" {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_DOMAIN_EMPTY", "en"))
	}

	if len(domainPart) > 253 {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_DOMAIN_TOO_LONG", "en"))
	}

	// Check for valid characters in domain
	for _, char := range domainPart {
		if !isValidDomainChar(char) {
			return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_DOMAIN_INVALID_CHARS", "en"))
		}
	}

	// Check for consecutive dots
	if strings.Contains(domainPart, "..") {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_DOMAIN_CONSECUTIVE_DOTS", "en"))
	}

	// Check for leading/trailing dots
	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_DOMAIN_LEADING_TRAILING_DOTS", "en"))
	}

	// Check for valid TLD (at least 2 characters)
	domainParts := strings.Split(domainPart, ".")
	if len(domainParts) < 2 {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_DOMAIN_INVALID_TLD", "en"))
	}

	tld := domainParts[len(domainParts)-1]
	if len(tld) < 2 {
		return errors.New(errors.ErrInvalidEmail, i18n.T("EMAIL_DOMAIN_INVALID_TLD", "en"))
	}

	return nil
}

// isValidLocalPartChar checks if character is valid in local part
func isValidLocalPartChar(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) ||
		char == '!' || char == '#' || char == '$' || char == '%' ||
		char == '&' || char == '\'' || char == '*' || char == '+' ||
		char == '-' || char == '/' || char == '=' || char == '?' ||
		char == '^' || char == '_' || char == '`' || char == '{' ||
		char == '|' || char == '}' || char == '~' || char == '.'
}

// isValidDomainChar checks if character is valid in domain
func isValidDomainChar(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || char == '-' || char == '.'
}

// containsSuspiciousPatterns checks for potentially malicious patterns
func containsSuspiciousPatterns(email string) bool {
	suspiciousPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"<iframe", "<object", "<embed", "data:text/html",
		"../../", "..\\", "file://", "ftp://", "gopher://",
	}

	emailLower := strings.ToLower(email)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(emailLower, pattern) {
			return true
		}
	}

	return false
}

// MustNewEmail creates a new Email value object and panics if validation fails
// Use only in tests or when you're certain the email is valid
func MustNewEmail(email string) Email {
	e, err := NewEmail(email)
	if err != nil {
		panic(err)
	}
	return e
}
