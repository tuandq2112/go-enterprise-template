package logger

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

// Level represents log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns string representation of log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a logger
type Logger struct {
	level Level
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	fatal *log.Logger
}

// NewLogger creates a new logger
func NewLogger(level Level) *Logger {
	flags := log.LstdFlags | log.Lshortfile

	return &Logger{
		level: level,
		debug: log.New(os.Stdout, "[DEBUG] ", flags),
		info:  log.New(os.Stdout, "[INFO] ", flags),
		warn:  log.New(os.Stderr, "[WARN] ", flags),
		error: log.New(os.Stderr, "[ERROR] ", flags),
		fatal: log.New(os.Stderr, "[FATAL] ", flags),
	}
}

// Debug logs debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= LevelDebug {
		l.debug.Printf(format, v...)
	}
}

// Info logs info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= LevelInfo {
		l.info.Printf(format, v...)
	}
}

// Warn logs warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= LevelWarn {
		l.warn.Printf(format, v...)
	}
}

// Error logs error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= LevelError {
		l.error.Printf(format, v...)
	}
}

// Fatal logs fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= LevelFatal {
		l.fatal.Printf(format, v...)
		os.Exit(1)
	}
}

// WithContext creates a logger with context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// In a more sophisticated implementation, you might extract
	// request ID, user ID, etc. from context
	return l
}

// WithFields creates a logger with fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// In a more sophisticated implementation, you might add
	// structured logging with fields
	return l
}

// HTTPMiddleware creates HTTP middleware for logging
func (l *Logger) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call the next handler
			next.ServeHTTP(wrappedWriter, r)

			// Log the request
			duration := time.Since(start)
			l.Info("%s %s %d %v", r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// Default logger instance
var defaultLogger = NewLogger(LevelInfo)

// SetLevel sets the default logger level
func SetLevel(level Level) {
	defaultLogger = NewLogger(level)
}

// Debug logs debug message using default logger
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info logs info message using default logger
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warn logs warning message using default logger
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Error logs error message using default logger
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Fatal logs fatal message using default logger
func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}
