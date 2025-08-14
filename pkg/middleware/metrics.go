package middleware

import (
	"net/http"
	"strconv"
	"time"

	"go-clean-ddd-es-template/pkg/metrics"
)

// MetricsMiddleware wraps HTTP handlers to collect metrics
type MetricsMiddleware struct {
	metrics *metrics.Metrics
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(m *metrics.Metrics) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: m,
	}
}

// Wrap wraps an HTTP handler with metrics collection
func (m *MetricsMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Record in-flight request
		m.metrics.RecordHTTPRequestInFlight(r.Method, r.URL.Path, 1)
		defer m.metrics.RecordHTTPRequestInFlight(r.Method, r.URL.Path, 0)

		// Call the next handler
		next.ServeHTTP(wrappedWriter, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrappedWriter.statusCode)
		m.metrics.RecordHTTPRequest(r.Method, r.URL.Path, status, duration)
	})
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
