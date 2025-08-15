package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"go-clean-ddd-es-template/pkg/errors"
	"go-clean-ddd-es-template/pkg/logger"
)

// ValidationConfig holds validation configuration
type ValidationConfig struct {
	MaxRequestSize    int64         // Maximum request body size in bytes
	MaxHeaderSize     int64         // Maximum header size in bytes
	RateLimitRequests int           // Number of requests per window
	RateLimitWindow   time.Duration // Time window for rate limiting
	AllowedMethods    []string      // Allowed HTTP methods
	BlockedPatterns   []string      // Patterns to block in requests
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		MaxHeaderSize:     1 * 1024 * 1024,  // 1MB
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
		AllowedMethods:    []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		BlockedPatterns: []string{
			"<script", "javascript:", "vbscript:", "onload=", "onerror=",
			"<iframe", "<object", "<embed", "data:text/html",
			"../../", "..\\", "file://", "ftp://", "gopher://",
			"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
			"UNION", "OR", "AND", "WHERE", "FROM", "JOIN",
		},
	}
}

// ValidationMiddleware provides input validation and security checks
type ValidationMiddleware struct {
	config *ValidationConfig
	logger logger.Logger
	// Rate limiting storage (in production, use Redis or similar)
	requestCounts map[string]int
	lastReset     time.Time
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(config *ValidationConfig, logger logger.Logger) *ValidationMiddleware {
	if config == nil {
		config = DefaultValidationConfig()
	}
	return &ValidationMiddleware{
		config:        config,
		logger:        logger,
		requestCounts: make(map[string]int),
		lastReset:     time.Now(),
	}
}

// ValidateRequest validates incoming HTTP requests
func (vm *ValidationMiddleware) ValidateRequest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check HTTP method
			if !vm.isMethodAllowed(r.Method) {
				vm.logger.Warn("Blocked request with disallowed method: %s", r.Method)
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Check request size
			if r.ContentLength > vm.config.MaxRequestSize {
				vm.logger.Warn("Request too large: %d bytes", r.ContentLength)
				http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Check headers
			if err := vm.validateHeaders(r); err != nil {
				vm.logger.Warn("Invalid headers: %v", err)
				http.Error(w, "Invalid headers", http.StatusBadRequest)
				return
			}

			// Rate limiting
			if err := vm.checkRateLimit(r); err != nil {
				vm.logger.Warn("Rate limit exceeded: %v", err)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Validate request body
			if err := vm.validateRequestBody(r); err != nil {
				vm.logger.Warn("Invalid request body: %v", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// isMethodAllowed checks if the HTTP method is allowed
func (vm *ValidationMiddleware) isMethodAllowed(method string) bool {
	for _, allowed := range vm.config.AllowedMethods {
		if strings.EqualFold(method, allowed) {
			return true
		}
	}
	return false
}

// validateHeaders validates request headers
func (vm *ValidationMiddleware) validateHeaders(r *http.Request) error {
	// Check header size
	var headerSize int64
	for name, values := range r.Header {
		headerSize += int64(len(name))
		for _, value := range values {
			headerSize += int64(len(value))
		}
	}

	if headerSize > vm.config.MaxHeaderSize {
		return errors.New(errors.ErrBadRequest, "Headers too large")
	}

	// Check for suspicious headers
	suspiciousHeaders := []string{
		"X-Forwarded-For", "X-Real-IP", "X-Forwarded-Host",
		"X-Forwarded-Proto", "X-Forwarded-Port",
	}

	for _, header := range suspiciousHeaders {
		if r.Header.Get(header) != "" {
			vm.logger.Warn("Suspicious header detected: %s", header)
		}
	}

	return nil
}

// checkRateLimit implements basic rate limiting
func (vm *ValidationMiddleware) checkRateLimit(r *http.Request) error {
	// Reset counter if window has passed
	if time.Since(vm.lastReset) > vm.config.RateLimitWindow {
		vm.requestCounts = make(map[string]int)
		vm.lastReset = time.Now()
	}

	// Get client identifier (IP address)
	clientIP := vm.getClientIP(r)

	// Check rate limit
	if vm.requestCounts[clientIP] >= vm.config.RateLimitRequests {
		return errors.New(errors.ErrBadRequest, "Rate limit exceeded")
	}

	// Increment counter
	vm.requestCounts[clientIP]++

	return nil
}

// validateRequestBody validates request body
func (vm *ValidationMiddleware) validateRequestBody(r *http.Request) error {
	// Only validate for methods that typically have bodies
	if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
		return nil
	}

	// Read body for validation
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.Wrap(err, errors.ErrBadRequest, "Failed to read request body")
	}

	// Restore body for next handlers
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Check for blocked patterns
	if vm.containsBlockedPatterns(string(body)) {
		return errors.New(errors.ErrBadRequest, "Request contains blocked patterns")
	}

	// Check for null bytes
	if bytes.Contains(body, []byte{0}) {
		return errors.New(errors.ErrBadRequest, "Request contains null bytes")
	}

	return nil
}

// containsBlockedPatterns checks if content contains blocked patterns
func (vm *ValidationMiddleware) containsBlockedPatterns(content string) bool {
	contentLower := strings.ToLower(content)
	for _, pattern := range vm.config.BlockedPatterns {
		patternLower := strings.ToLower(pattern)

		// For SQL keywords, check for word boundaries
		if isSQLKeyword(pattern) {
			if hasWordBoundary(contentLower, patternLower) {
				return true
			}
		} else {
			// For other patterns, use simple contains
			if strings.Contains(contentLower, patternLower) {
				return true
			}
		}
	}
	return false
}

// isSQLKeyword checks if a pattern is a SQL keyword
func isSQLKeyword(pattern string) bool {
	sqlKeywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "UNION", "OR", "AND", "WHERE", "FROM", "JOIN"}
	patternUpper := strings.ToUpper(pattern)
	for _, keyword := range sqlKeywords {
		if patternUpper == keyword {
			return true
		}
	}
	return false
}

// hasWordBoundary checks if a word appears with proper boundaries
func hasWordBoundary(content, word string) bool {
	// Split content into words
	words := strings.Fields(content)
	for _, w := range words {
		// Remove common punctuation
		w = strings.Trim(w, ".,;:!?\"'()[]{}")
		if strings.EqualFold(w, word) {
			return true
		}
	}
	return false
}

// getClientIP extracts the real client IP address
func (vm *ValidationMiddleware) getClientIP(r *http.Request) string {
	// Check for forwarded headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIndex := strings.Index(ip, ","); commaIndex != -1 {
			return strings.TrimSpace(ip[:commaIndex])
		}
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fallback to remote address
	if r.RemoteAddr != "" {
		// Remove port if present
		if colonIndex := strings.LastIndex(r.RemoteAddr, ":"); colonIndex != -1 {
			return r.RemoteAddr[:colonIndex]
		}
		return r.RemoteAddr
	}

	return "unknown"
}

// SanitizeInput sanitizes input strings
func (vm *ValidationMiddleware) SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters (except newlines and tabs)
	var sanitized strings.Builder
	for _, char := range input {
		if char == '\n' || char == '\t' || char == '\r' {
			sanitized.WriteRune(char)
		} else if !isControlChar(char) {
			sanitized.WriteRune(char)
		}
	}

	return sanitized.String()
}

// isControlChar checks if a character is a control character
func isControlChar(char rune) bool {
	return char < 32 || (char >= 127 && char < 160)
}
