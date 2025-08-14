package middleware

import (
	"net/http"
	"time"

	"go-clean-ddd-es-template/pkg/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates a middleware that adds tracing to HTTP requests
func TracingMiddleware(tracer *tracing.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tracer == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Start span for the request
			ctx, span := tracer.StartSpan(r.Context(), "http.request",
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()

			// Add request ID to context
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			span.SetAttributes(attribute.String("request.id", requestID))

			// Create response writer wrapper to capture status code
			wrappedWriter := &tracingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Record start time
			start := time.Now()

			// Call next handler
			next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

			// Record end time and duration
			duration := time.Since(start)
			span.SetAttributes(
				attribute.Int("http.status_code", wrappedWriter.statusCode),
				attribute.Int64("http.duration_ms", duration.Milliseconds()),
			)

			// Add events for request start and end
			span.AddEvent("request.start", trace.WithAttributes(
				attribute.String("request.id", requestID),
			))
			span.AddEvent("request.end", trace.WithAttributes(
				attribute.String("request.id", requestID),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			))
		})
	}
}

// tracingResponseWriter wraps http.ResponseWriter to capture status code
type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *tracingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *tracingResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
