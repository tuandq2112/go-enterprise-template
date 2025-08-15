package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	failureThreshold int           // Number of failures before opening circuit
	timeout          time.Duration // Time to wait before trying half-open
	successThreshold int           // Number of successes to close circuit

	// State
	state       CircuitState
	failures    int
	successes   int
	lastFailure time.Time
	lastSuccess time.Time

	// Metrics
	totalRequests   int64
	totalFailures   int64
	totalSuccesses  int64
	lastStateChange time.Time
}

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	Timeout          time.Duration `json:"timeout"`
	SuccessThreshold int           `json:"success_threshold"`
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		Timeout:          30 * time.Second,
		SuccessThreshold: 3,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: config.FailureThreshold,
		timeout:          config.Timeout,
		successThreshold: config.SuccessThreshold,
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.beforeExecution(); err != nil {
		return err
	}

	err := fn()
	cb.afterExecution(err)
	return err
}

// ExecuteWithResult runs a function that returns a result with circuit breaker protection
func (cb *CircuitBreaker) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	if err := cb.beforeExecution(); err != nil {
		return nil, err
	}

	result, err := fn()
	cb.afterExecution(err)
	return result, err
}

// beforeExecution checks if circuit breaker allows execution
func (cb *CircuitBreaker) beforeExecution() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++

	switch cb.state {
	case StateClosed:
		return nil // Allow execution

	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.timeout {
			// Timeout reached, try half-open
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
			cb.successes = 0
			return nil
		}
		return fmt.Errorf("circuit breaker is OPEN: %w", ErrCircuitOpen)

	case StateHalfOpen:
		return nil // Allow execution to test if service is back

	default:
		return fmt.Errorf("unknown circuit breaker state: %v", cb.state)
	}
}

// afterExecution updates circuit breaker state based on execution result
func (cb *CircuitBreaker) afterExecution(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure handles failure and updates circuit breaker state
func (cb *CircuitBreaker) recordFailure() {
	cb.totalFailures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.failureThreshold {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}

	case StateHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		cb.failures = cb.failureThreshold // Ensure it stays open
	}
}

// recordSuccess handles success and updates circuit breaker state
func (cb *CircuitBreaker) recordSuccess() {
	cb.totalSuccesses++
	cb.lastSuccess = time.Now()

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.successThreshold {
			// Enough successes, close the circuit
			cb.state = StateClosed
			cb.lastStateChange = time.Now()
			cb.failures = 0
			cb.successes = 0
		}
	}
}

// GetState returns current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:            cb.state,
		Failures:         cb.failures,
		Successes:        cb.successes,
		TotalRequests:    cb.totalRequests,
		TotalFailures:    cb.totalFailures,
		TotalSuccesses:   cb.totalSuccesses,
		LastFailure:      cb.lastFailure,
		LastSuccess:      cb.lastSuccess,
		LastStateChange:  cb.lastStateChange,
		FailureThreshold: cb.failureThreshold,
		SuccessThreshold: cb.successThreshold,
		Timeout:          cb.timeout,
	}
}

// CircuitBreakerStats holds statistics for circuit breaker
type CircuitBreakerStats struct {
	State            CircuitState  `json:"state"`
	Failures         int           `json:"failures"`
	Successes        int           `json:"successes"`
	TotalRequests    int64         `json:"total_requests"`
	TotalFailures    int64         `json:"total_failures"`
	TotalSuccesses   int64         `json:"total_successes"`
	LastFailure      time.Time     `json:"last_failure"`
	LastSuccess      time.Time     `json:"last_success"`
	LastStateChange  time.Time     `json:"last_state_change"`
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateOpen
	cb.lastStateChange = time.Now()
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.lastStateChange = time.Now()
	cb.failures = 0
	cb.successes = 0
}

// Reset resets all circuit breaker statistics
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.successes = 0
	cb.totalRequests = 0
	cb.totalFailures = 0
	cb.totalSuccesses = 0
	cb.lastFailure = time.Time{}
	cb.lastSuccess = time.Time{}
	cb.lastStateChange = time.Now()
}

// Errors
var (
	ErrCircuitOpen = fmt.Errorf("circuit breaker is open")
)
