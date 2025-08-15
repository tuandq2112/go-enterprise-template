package repositories

import (
	"context"
	"time"

	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/domain/repositories"
	"go-clean-ddd-es-template/pkg/resilience"
)

// CircuitBreakerEventPublisher wraps EventPublisher with circuit breaker
type CircuitBreakerEventPublisher struct {
	publisher      repositories.EventPublisher
	circuitBreaker *resilience.CircuitBreaker
}

// NewCircuitBreakerEventPublisher creates a new circuit breaker event publisher
func NewCircuitBreakerEventPublisher(publisher repositories.EventPublisher, config resilience.CircuitBreakerConfig) *CircuitBreakerEventPublisher {
	return &CircuitBreakerEventPublisher{
		publisher:      publisher,
		circuitBreaker: resilience.NewCircuitBreaker(config),
	}
}

// PublishEvent wraps publisher.PublishEvent with circuit breaker
func (cb *CircuitBreakerEventPublisher) PublishEvent(ctx context.Context, event *events.Event) error {
	_, err := cb.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return nil, cb.publisher.PublishEvent(ctx, event)
	})
	return err
}

// PublishEvents wraps publisher.PublishEvents with circuit breaker
func (cb *CircuitBreakerEventPublisher) PublishEvents(ctx context.Context, events []*events.Event) error {
	_, err := cb.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return nil, cb.publisher.PublishEvents(ctx, events)
	})
	return err
}

// PublishEventWithRetry publishes an event with retry logic and circuit breaker
func (cb *CircuitBreakerEventPublisher) PublishEventWithRetry(ctx context.Context, event *events.Event, maxRetries int) error {
	return cb.circuitBreaker.Execute(ctx, func() error {
		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			if err := cb.publisher.PublishEvent(ctx, event); err == nil {
				return nil
			} else {
				lastErr = err
				if attempt < maxRetries {
					// Exponential backoff
					backoff := time.Duration(attempt) * time.Second
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(backoff):
						continue
					}
				}
			}
		}
		return lastErr
	})
}

// PublishEventsWithRetry publishes multiple events with retry logic and circuit breaker
func (cb *CircuitBreakerEventPublisher) PublishEventsWithRetry(ctx context.Context, events []*events.Event, maxRetries int) error {
	return cb.circuitBreaker.Execute(ctx, func() error {
		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			if err := cb.publisher.PublishEvents(ctx, events); err == nil {
				return nil
			} else {
				lastErr = err
				if attempt < maxRetries {
					// Exponential backoff
					backoff := time.Duration(attempt) * time.Second
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(backoff):
						continue
					}
				}
			}
		}
		return lastErr
	})
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreakerEventPublisher) GetStats() resilience.CircuitBreakerStats {
	return cb.circuitBreaker.GetStats()
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreakerEventPublisher) ForceOpen() {
	cb.circuitBreaker.ForceOpen()
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreakerEventPublisher) ForceClose() {
	cb.circuitBreaker.ForceClose()
}
