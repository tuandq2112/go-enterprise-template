package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZapLogger(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		format string
		want   bool
	}{
		{
			name:   "valid debug level with json format",
			level:  "debug",
			format: "json",
			want:   true,
		},
		{
			name:   "valid info level with text format",
			level:  "info",
			format: "text",
			want:   true,
		},
		{
			name:   "valid warn level with json format",
			level:  "warn",
			format: "json",
			want:   true,
		},
		{
			name:   "valid error level with text format",
			level:  "error",
			format: "text",
			want:   true,
		},
		{
			name:   "valid fatal level with json format",
			level:  "fatal",
			format: "json",
			want:   true,
		},
		{
			name:   "invalid level defaults to info",
			level:  "invalid",
			format: "text",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewZapLogger(tt.level, tt.format)
			if tt.want {
				require.NoError(t, err)
				assert.NotNil(t, logger)
				assert.Implements(t, (*Logger)(nil), logger)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestZapLogger_Logging(t *testing.T) {
	logger, err := NewZapLogger("debug", "text")
	require.NoError(t, err)

	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
	// Don't test Fatal as it calls os.Exit(1)
}

func TestZapLogger_WithContext(t *testing.T) {
	logger, err := NewZapLogger("info", "text")
	require.NoError(t, err)

	// Test with context that has request_id
	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	loggerWithContext := logger.WithContext(ctx)
	assert.NotNil(t, loggerWithContext)
	assert.Implements(t, (*Logger)(nil), loggerWithContext)

	// Test with context without request_id
	ctxNoID := context.Background()
	loggerWithContextNoID := logger.WithContext(ctxNoID)
	assert.NotNil(t, loggerWithContextNoID)
	assert.Implements(t, (*Logger)(nil), loggerWithContextNoID)
}

func TestZapLogger_WithFields(t *testing.T) {
	logger, err := NewZapLogger("info", "text")
	require.NoError(t, err)

	fields := map[string]interface{}{
		"user_id": "123",
		"action":  "test",
		"count":   42,
	}

	loggerWithFields := logger.WithFields(fields)
	assert.NotNil(t, loggerWithFields)
	assert.Implements(t, (*Logger)(nil), loggerWithFields)

	// Test with empty fields
	loggerWithEmptyFields := logger.WithFields(map[string]interface{}{})
	assert.NotNil(t, loggerWithEmptyFields)
	assert.Implements(t, (*Logger)(nil), loggerWithEmptyFields)
}

func TestZapLogger_HTTPMiddleware(t *testing.T) {
	logger, err := NewZapLogger("info", "text")
	require.NoError(t, err)

	middleware := logger.HTTPMiddleware()
	assert.NotNil(t, middleware)

	// Test that middleware is a function that returns http.Handler
	handler := middleware(nil)
	assert.NotNil(t, handler)
}

func TestZapLogger_GetZapLogger(t *testing.T) {
	logger, err := NewZapLogger("info", "text")
	require.NoError(t, err)

	zapLogger := logger.GetZapLogger()
	assert.NotNil(t, zapLogger)
}

func TestZapLogger_GetZapSugar(t *testing.T) {
	logger, err := NewZapLogger("info", "text")
	require.NoError(t, err)

	zapSugar := logger.GetZapSugar()
	assert.NotNil(t, zapSugar)
}

func TestZapLogger_Sync(t *testing.T) {
	logger, err := NewZapLogger("info", "text")
	require.NoError(t, err)

	// Sync might return an error on some systems (like stdout sync)
	// We'll just test that it doesn't panic
	err = logger.Sync()
	// Don't assert on error as it can fail on some systems
}
