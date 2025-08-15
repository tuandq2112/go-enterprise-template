package logger

import (
	"context"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger implements Logger interface using Zap
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewZapLogger creates a new Zap-based logger
func NewZapLogger(level string, format string) (*ZapLogger, error) {
	// Parse log level
	zapLevel, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	// Configure encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller())
	sugar := logger.Sugar()

	return &ZapLogger{
		logger: logger,
		sugar:  sugar,
	}, nil
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

// Debug logs debug message
func (l *ZapLogger) Debug(format string, v ...interface{}) {
	l.sugar.Debugf(format, v...)
}

// Info logs info message
func (l *ZapLogger) Info(format string, v ...interface{}) {
	l.sugar.Infof(format, v...)
}

// Warn logs warning message
func (l *ZapLogger) Warn(format string, v ...interface{}) {
	l.sugar.Warnf(format, v...)
}

// Error logs error message
func (l *ZapLogger) Error(format string, v ...interface{}) {
	l.sugar.Errorf(format, v...)
}

// Fatal logs fatal message and exits
func (l *ZapLogger) Fatal(format string, v ...interface{}) {
	l.sugar.Fatalf(format, v...)
}

// WithContext creates a logger with context
func (l *ZapLogger) WithContext(ctx context.Context) Logger {
	// Extract request ID from context if available
	if requestID, ok := ctx.Value("request_id").(string); ok {
		newLogger := l.logger.With(zap.String("request_id", requestID))
		return &ZapLogger{
			logger: newLogger,
			sugar:  newLogger.Sugar(),
		}
	}
	return l
}

// WithFields creates a logger with fields
func (l *ZapLogger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	newLogger := l.logger.With(zapFields...)
	return &ZapLogger{
		logger: newLogger,
		sugar:  newLogger.Sugar(),
	}
}

// HTTPMiddleware creates HTTP middleware for logging
func (l *ZapLogger) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrappedWriter := &zapResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call the next handler
			next.ServeHTTP(wrappedWriter, r)

			// Log the request
			duration := time.Since(start)
			l.Info("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrappedWriter.statusCode,
				"duration", duration,
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// Sync flushes any buffered log entries
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// GetZapLogger returns the underlying zap logger
func (l *ZapLogger) GetZapLogger() *zap.Logger {
	return l.logger
}

// GetZapSugar returns the underlying zap sugared logger
func (l *ZapLogger) GetZapSugar() *zap.SugaredLogger {
	return l.sugar
}

// zapResponseWriter wraps http.ResponseWriter to capture status code
type zapResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *zapResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *zapResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
