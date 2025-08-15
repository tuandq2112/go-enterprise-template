package logger

import (
	"context"
	"net/http"
)

// Logger interface defines the contract for logging
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
	WithContext(ctx context.Context) Logger
	WithFields(fields map[string]interface{}) Logger
	HTTPMiddleware() func(http.Handler) http.Handler
}

// NewLoggerFromConfig creates a new logger based on configuration
func NewLoggerFromConfig(level string, format string) (Logger, error) {
	// Use Zap logger by default for better performance and structured logging
	return NewZapLogger(level, format)
}
