package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCircuitBreaker(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 3,
		Timeout:          30 * time.Second,
		SuccessThreshold: 2,
	}

	cb := NewCircuitBreaker(config)

	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 3, cb.GetStats().FailureThreshold)
	assert.Equal(t, 30*time.Second, cb.GetStats().Timeout)
	assert.Equal(t, 2, cb.GetStats().SuccessThreshold)
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, int64(1), cb.GetStats().TotalRequests)
	assert.Equal(t, int64(1), cb.GetStats().TotalSuccesses)
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          100 * time.Millisecond,
		SuccessThreshold: 1,
	})

	// First failure
	err := cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	assert.Error(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 1, cb.GetStats().Failures)

	// Second failure - should open circuit
	err = cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 2, cb.GetStats().Failures)
}

func TestCircuitBreaker_Execute_CircuitOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		Timeout:          100 * time.Millisecond,
		SuccessThreshold: 1,
	})

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	// Try to execute while circuit is open
	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is OPEN")
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_Execute_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		Timeout:          50 * time.Millisecond,
		SuccessThreshold: 1,
	})

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	// Wait for timeout to go to half-open
	time.Sleep(100 * time.Millisecond)

	// Should be in half-open state
	assert.Equal(t, StateHalfOpen, cb.GetState())

	// Success in half-open should close circuit
	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetStats().Failures)
	assert.Equal(t, 1, cb.GetStats().Successes)
}

func TestCircuitBreaker_Execute_HalfOpen_Failure(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		Timeout:          50 * time.Millisecond,
		SuccessThreshold: 1,
	})

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	// Wait for timeout to go to half-open
	time.Sleep(100 * time.Millisecond)

	// Failure in half-open should open circuit again
	err := cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_ExecuteWithResult(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	result, err := cb.ExecuteWithResult(context.Background(), func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_ExecuteWithResult_Failure(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		Timeout:          100 * time.Millisecond,
		SuccessThreshold: 1,
	})

	result, err := cb.ExecuteWithResult(context.Background(), func() (interface{}, error) {
		return nil, errors.New("test error")
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_ForceOpen(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	cb.ForceOpen()

	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_ForceClose(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		Timeout:          100 * time.Millisecond,
		SuccessThreshold: 1,
	})

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	cb.ForceClose()

	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetStats().Failures)
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	// Execute some operations
	cb.Execute(context.Background(), func() error {
		return errors.New("test error")
	})

	cb.Reset()

	stats := cb.GetStats()
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.TotalFailures)
	assert.Equal(t, int64(0), stats.TotalSuccesses)
}

func TestCircuitState_String(t *testing.T) {
	assert.Equal(t, "CLOSED", StateClosed.String())
	assert.Equal(t, "OPEN", StateOpen.String())
	assert.Equal(t, "HALF_OPEN", StateHalfOpen.String())
}
