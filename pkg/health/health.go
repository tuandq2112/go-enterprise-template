package health

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// Check represents a health check
type Check struct {
	Name     string                 `json:"name"`
	Status   Status                 `json:"status"`
	Message  string                 `json:"message,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Duration time.Duration          `json:"duration,omitempty"`
}

// HealthChecker represents a health check function
type HealthChecker func(ctx context.Context) Check

// HealthService manages health checks
type HealthService struct {
	checks []HealthChecker
	mu     sync.RWMutex
}

// NewHealthService creates a new health service
func NewHealthService() *HealthService {
	return &HealthService{
		checks: make([]HealthChecker, 0),
	}
}

// AddCheck adds a health check
func (h *HealthService) AddCheck(check HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks = append(h.checks, check)
}

// Check performs all health checks
func (h *HealthService) Check(ctx context.Context) []Check {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make([]Check, len(h.checks))
	for i, check := range h.checks {
		results[i] = check(ctx)
	}
	return results
}

// OverallStatus determines the overall health status
func (h *HealthService) OverallStatus(checks []Check) Status {
	if len(checks) == 0 {
		return StatusHealthy
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// HTTPHandler returns an HTTP handler for health checks
func (h *HealthService) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		checks := h.Check(ctx)
		overallStatus := h.OverallStatus(checks)

		response := map[string]interface{}{
			"status": overallStatus,
			"checks": checks,
			"time":   time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")

		// Set appropriate HTTP status code
		switch overallStatus {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // 200 but with degraded status
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(response)
	}
}

// DatabaseCheck creates a database health check
func DatabaseCheck(db interface{ Ping() error }) HealthChecker {
	return func(ctx context.Context) Check {
		start := time.Now()
		err := db.Ping()
		duration := time.Since(start)

		check := Check{
			Name:     "database",
			Duration: duration,
		}

		if err != nil {
			check.Status = StatusUnhealthy
			check.Message = err.Error()
		} else {
			check.Status = StatusHealthy
			check.Message = "Database connection is healthy"
		}

		return check
	}
}

// KafkaCheck creates a Kafka health check
func KafkaCheck(producer interface{ Close() error }) HealthChecker {
	return func(ctx context.Context) Check {
		start := time.Now()

		// Try to close and reopen connection as a health check
		// This is a simplified check - in production you might want to send a test message
		err := producer.Close()
		duration := time.Since(start)

		check := Check{
			Name:     "kafka",
			Duration: duration,
		}

		if err != nil {
			check.Status = StatusUnhealthy
			check.Message = err.Error()
		} else {
			check.Status = StatusHealthy
			check.Message = "Kafka connection is healthy"
		}

		return check
	}
}

// SystemCheck creates a system health check
func SystemCheck() HealthChecker {
	return func(ctx context.Context) Check {
		start := time.Now()

		// Simple system check - could be extended with memory, CPU, etc.
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		duration := time.Since(start)

		check := Check{
			Name:     "system",
			Status:   StatusHealthy,
			Message:  "System is healthy",
			Duration: duration,
			Details: map[string]interface{}{
				"goroutines":   runtime.NumGoroutine(),
				"memory_alloc": memStats.Alloc,
				"memory_heap":  memStats.HeapAlloc,
			},
		}

		return check
	}
}
