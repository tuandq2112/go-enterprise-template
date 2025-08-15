package messagebroker

import (
	"context"
	"fmt"
	"time"

	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/pkg/resilience"

	"github.com/IBM/sarama"
)

// CircuitBreakerMessageBroker wraps MessageBroker with circuit breaker
type CircuitBreakerMessageBroker struct {
	broker         MessageBroker
	circuitBreaker *resilience.CircuitBreaker
	config         *config.MessageBrokerConfig
}

// NewCircuitBreakerMessageBroker creates a new circuit breaker message broker
func NewCircuitBreakerMessageBroker(broker MessageBroker, config *config.MessageBrokerConfig, cbConfig resilience.CircuitBreakerConfig) *CircuitBreakerMessageBroker {
	return &CircuitBreakerMessageBroker{
		broker:         broker,
		circuitBreaker: resilience.NewCircuitBreaker(cbConfig),
		config:         config,
	}
}

// Connect wraps broker.Connect with circuit breaker
func (cb *CircuitBreakerMessageBroker) Connect() error {
	_, err := cb.circuitBreaker.ExecuteWithResult(context.Background(), func() (interface{}, error) {
		return nil, cb.broker.Connect()
	})
	return err
}

// Close wraps broker.Close with circuit breaker
func (cb *CircuitBreakerMessageBroker) Close() error {
	_, err := cb.circuitBreaker.ExecuteWithResult(context.Background(), func() (interface{}, error) {
		return nil, cb.broker.Close()
	})
	return err
}

// Publish wraps broker.Publish with circuit breaker
func (cb *CircuitBreakerMessageBroker) Publish(topic string, message []byte) error {
	_, err := cb.circuitBreaker.ExecuteWithResult(context.Background(), func() (interface{}, error) {
		return nil, cb.broker.Publish(topic, message)
	})
	return err
}

// Subscribe wraps broker.Subscribe with circuit breaker
func (cb *CircuitBreakerMessageBroker) Subscribe(topic string, handler func([]byte)) error {
	_, err := cb.circuitBreaker.ExecuteWithResult(context.Background(), func() (interface{}, error) {
		return nil, cb.broker.Subscribe(topic, handler)
	})
	return err
}

// GetConsumer wraps broker.GetConsumer with circuit breaker
func (cb *CircuitBreakerMessageBroker) GetConsumer() sarama.Consumer {
	// GetConsumer doesn't need circuit breaker as it's just returning a reference
	return cb.broker.GetConsumer()
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreakerMessageBroker) GetStats() resilience.CircuitBreakerStats {
	return cb.circuitBreaker.GetStats()
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreakerMessageBroker) ForceOpen() {
	cb.circuitBreaker.ForceOpen()
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreakerMessageBroker) ForceClose() {
	cb.circuitBreaker.ForceClose()
}

// PublishWithRetry publishes a message with retry logic and circuit breaker
func (cb *CircuitBreakerMessageBroker) PublishWithRetry(ctx context.Context, topic string, message []byte, maxRetries int) error {
	return cb.circuitBreaker.Execute(ctx, func() error {
		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			if err := cb.broker.Publish(topic, message); err == nil {
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
		return fmt.Errorf("failed to publish after %d attempts: %w", maxRetries, lastErr)
	})
}
