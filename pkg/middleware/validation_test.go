package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-clean-ddd-es-template/pkg/logger"
)

func TestValidationMiddleware_ValidateRequest(t *testing.T) {
	// Create test logger
	testLogger, _ := logger.NewLoggerFromConfig("info", "text")

	// Create validation middleware
	config := DefaultValidationConfig()
	config.RateLimitRequests = 5 // Lower for testing
	vm := NewValidationMiddleware(config, testLogger)

	tests := []struct {
		name           string
		method         string
		body           string
		headers        map[string]string
		expectedStatus int
		description    string
	}{
		{
			name:           "valid GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			description:    "Should allow valid GET request",
		},
		{
			name:           "valid POST request",
			method:         "POST",
			body:           `{"name": "John", "email": "john@example.com"}`,
			expectedStatus: http.StatusOK,
			description:    "Should allow valid POST request",
		},
		{
			name:           "disallowed method",
			method:         "TRACE",
			expectedStatus: http.StatusMethodNotAllowed,
			description:    "Should block disallowed HTTP method",
		},
		{
			name:           "request with blocked pattern",
			method:         "POST",
			body:           `{"name": "<script>alert('xss')</script>"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Should block request with XSS pattern",
		},
		{
			name:           "request with SQL injection",
			method:         "POST",
			body:           `{"query": "SELECT * FROM users WHERE id = 1 OR 1=1"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Should block request with SQL injection pattern",
		},
		{
			name:           "request with null bytes",
			method:         "POST",
			body:           "{\"name\": \"John\x00Doe\"}",
			expectedStatus: http.StatusBadRequest,
			description:    "Should block request with null bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/test", body)

			// Add headers
			if tt.headers != nil {
				for key, value := range tt.headers {
					req.Header.Set(key, value)
				}
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Apply validation middleware
			middleware := vm.ValidateRequest()
			middleware(handler).ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.description, tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestValidationMiddleware_RateLimit(t *testing.T) {
	// Create test logger
	testLogger, _ := logger.NewLoggerFromConfig("info", "text")

	// Create validation middleware with low rate limit
	config := DefaultValidationConfig()
	config.RateLimitRequests = 2
	config.RateLimitWindow = time.Second
	vm := NewValidationMiddleware(config, testLogger)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	// Create response recorder
	rr := httptest.NewRecorder()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply validation middleware
	middleware := vm.ValidateRequest()

	// First request should succeed
	middleware(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("First request should succeed, got status %d", rr.Code)
	}

	// Second request should succeed
	rr = httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Second request should succeed, got status %d", rr.Code)
	}

	// Third request should be rate limited
	rr = httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should be rate limited, got status %d", rr.Code)
	}
}

func TestValidationMiddleware_SanitizeInput(t *testing.T) {
	// Create test logger
	testLogger, _ := logger.NewLoggerFromConfig("info", "text")

	// Create validation middleware
	vm := NewValidationMiddleware(DefaultValidationConfig(), testLogger)

	tests := []struct {
		input    string
		expected string
		name     string
	}{
		{
			input:    "Hello World",
			expected: "Hello World",
			name:     "normal string",
		},
		{
			input:    "Hello\x00World",
			expected: "HelloWorld",
			name:     "string with null bytes",
		},
		{
			input:    "Hello\x01World",
			expected: "HelloWorld",
			name:     "string with control characters",
		},
		{
			input:    "Hello\nWorld",
			expected: "Hello\nWorld",
			name:     "string with newline",
		},
		{
			input:    "Hello\tWorld",
			expected: "Hello\tWorld",
			name:     "string with tab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidationMiddleware_GetClientIP(t *testing.T) {
	// Create test logger
	testLogger, _ := logger.NewLoggerFromConfig("info", "text")

	// Create validation middleware
	vm := NewValidationMiddleware(DefaultValidationConfig(), testLogger)

	tests := []struct {
		headers    map[string]string
		remoteAddr string
		expected   string
		name       string
	}{
		{
			headers:  map[string]string{"X-Forwarded-For": "192.168.1.1"},
			expected: "192.168.1.1",
			name:     "X-Forwarded-For header",
		},
		{
			headers:  map[string]string{"X-Real-IP": "192.168.1.2"},
			expected: "192.168.1.2",
			name:     "X-Real-IP header",
		},
		{
			headers:  map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.1"},
			expected: "192.168.1.1",
			name:     "X-Forwarded-For with multiple IPs",
		},
		{
			remoteAddr: "192.168.1.3:12345",
			expected:   "192.168.1.3",
			name:       "RemoteAddr with port",
		},
		{
			remoteAddr: "192.168.1.4",
			expected:   "192.168.1.4",
			name:       "RemoteAddr without port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Set remote address
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			result := vm.getClientIP(req)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidationConfig_DefaultConfig(t *testing.T) {
	config := DefaultValidationConfig()

	// Check default values
	if config.MaxRequestSize != 10*1024*1024 {
		t.Errorf("expected MaxRequestSize %d, got %d", 10*1024*1024, config.MaxRequestSize)
	}

	if config.MaxHeaderSize != 1*1024*1024 {
		t.Errorf("expected MaxHeaderSize %d, got %d", 1*1024*1024, config.MaxHeaderSize)
	}

	if config.RateLimitRequests != 100 {
		t.Errorf("expected RateLimitRequests %d, got %d", 100, config.RateLimitRequests)
	}

	if config.RateLimitWindow != time.Minute {
		t.Errorf("expected RateLimitWindow %v, got %v", time.Minute, config.RateLimitWindow)
	}

	// Check that blocked patterns are set
	if len(config.BlockedPatterns) == 0 {
		t.Error("expected blocked patterns to be set")
	}
}

func TestValidationMiddleware_ContainsBlockedPatterns(t *testing.T) {
	// Create test logger
	testLogger, _ := logger.NewLoggerFromConfig("info", "text")

	// Create validation middleware
	vm := NewValidationMiddleware(DefaultValidationConfig(), testLogger)

	tests := []struct {
		content  string
		expected bool
		name     string
	}{
		{
			content:  "Hello World",
			expected: false,
			name:     "normal content",
		},
		{
			content:  "<script>alert('xss')</script>",
			expected: true,
			name:     "XSS script tag",
		},
		{
			content:  "javascript:alert('xss')",
			expected: true,
			name:     "javascript protocol",
		},
		{
			content:  "SELECT * FROM users",
			expected: true,
			name:     "SQL injection",
		},
		{
			content:  "file:///etc/passwd",
			expected: true,
			name:     "file protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.containsBlockedPatterns(tt.content)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for content: %s", tt.expected, result, tt.content)
			}
		})
	}
}
