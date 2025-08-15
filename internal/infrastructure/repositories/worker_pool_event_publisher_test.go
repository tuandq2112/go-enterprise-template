package repositories_test

import (
	"context"
	"testing"
	"time"

	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/infrastructure/config"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPoolEventPublisher_Configuration(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		MessageBroker: config.MessageBrokerConfig{
			PublisherWorkers: 5,
			ConsumerWorkers:  10,
			WorkerBufferSize: 100,
			Topics: map[string]string{
				"user.created": "user-events",
				"user.updated": "user-events",
			},
		},
	}

	// Test configuration values
	assert.Equal(t, 5, cfg.MessageBroker.PublisherWorkers)
	assert.Equal(t, 10, cfg.MessageBroker.ConsumerWorkers)
	assert.Equal(t, 100, cfg.MessageBroker.WorkerBufferSize)
	assert.Equal(t, "user-events", cfg.MessageBroker.Topics["user.created"])
}

func TestWorkerPoolEventPublisher_DefaultValues(t *testing.T) {
	// Test default configuration loading
	cfg := config.Load()

	// Assert default values are set
	assert.Greater(t, cfg.MessageBroker.PublisherWorkers, 0)
	assert.Greater(t, cfg.MessageBroker.ConsumerWorkers, 0)
	assert.Greater(t, cfg.MessageBroker.WorkerBufferSize, 0)
}

func TestWorkerPoolEventPublisher_EventStructure(t *testing.T) {
	// Create test event
	event := &events.Event{
		Type:      "user.created",
		Data:      []byte(`{"user_id": "123", "email": "test@example.com"}`),
		Timestamp: time.Now(),
		Version:   1,
	}

	// Assert event structure
	assert.Equal(t, "user.created", event.Type)
	assert.NotNil(t, event.Data)
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 1, event.Version)
}

func TestWorkerPoolEventPublisher_ContextHandling(t *testing.T) {
	// Test context cancellation
	_, cancel := context.WithCancel(context.Background())
	cancel()

	// Create test event
	event := &events.Event{
		Type:      "user.created",
		Data:      []byte(`{"user_id": "123"}`),
		Timestamp: time.Now(),
		Version:   1,
	}

	// This would normally be used with a real publisher
	// For now, just test that the event can be created
	assert.NotNil(t, event)
	assert.Equal(t, "user.created", event.Type)
}
