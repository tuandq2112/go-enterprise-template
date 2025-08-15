package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/pkg/resilience"
)

// EventConsumer handles event consumption with dead letter queue
type EventConsumer struct {
	eventHandlers   map[string]EventHandler
	deadLetterQueue *resilience.DeadLetterQueue
	logger          Logger
}

// EventHandler interface for handling specific event types
type EventHandler interface {
	HandleEvent(ctx context.Context, event *entities.UserEvent) error
}

// Logger interface for logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// EventConsumerConfig holds configuration for event consumer
type EventConsumerConfig struct {
	DLQConfig resilience.DeadLetterQueueConfig
}

// DefaultEventConsumerConfig returns default configuration
func DefaultEventConsumerConfig() EventConsumerConfig {
	return EventConsumerConfig{
		DLQConfig: resilience.DefaultDeadLetterQueueConfig(),
	}
}

// NewEventConsumer creates a new event consumer with dead letter queue
func NewEventConsumer(config EventConsumerConfig, logger Logger) *EventConsumer {
	// Create dead letter queue with in-memory storage for now
	dlq := resilience.NewDeadLetterQueue(config.DLQConfig, nil, nil)

	return &EventConsumer{
		eventHandlers:   make(map[string]EventHandler),
		deadLetterQueue: dlq,
		logger:          logger,
	}
}

// RegisterHandler registers an event handler for a specific event type
func (ec *EventConsumer) RegisterHandler(eventType string, handler EventHandler) {
	ec.eventHandlers[eventType] = handler
}

// HandleMessage processes a message with dead letter queue
func (ec *EventConsumer) HandleMessage(ctx context.Context, message []byte) error {
	// Parse event from message broker format
	var event events.Event
	if err := json.Unmarshal(message, &event); err != nil {
		ec.logger.Error("Failed to unmarshal event: %v", err)
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Convert to UserEvent format for processing
	userEvent := &entities.UserEvent{
		UserID:    "", // Will be extracted from event data
		EventType: event.Type,
		EventData: make(map[string]interface{}),
		Timestamp: event.Timestamp,
		Version:   event.Version,
	}

	// Parse event data
	if len(event.Data) > 0 {
		if err := json.Unmarshal(event.Data, &userEvent.EventData); err != nil {
			ec.logger.Error("Failed to unmarshal event data: %v", err)
			return fmt.Errorf("failed to unmarshal event data: %w", err)
		}
	}

	// Extract user_id from event data
	if userID, ok := userEvent.EventData["user_id"].(string); ok {
		userEvent.UserID = userID
	}

	// Process the event
	err := ec.processEvent(ctx, userEvent)
	if err != nil {
		// If processing failed, add to dead letter queue
		eventData := map[string]interface{}{
			"user_id":    userEvent.UserID,
			"event_type": userEvent.EventType,
			"event_data": userEvent.EventData,
			"timestamp":  userEvent.Timestamp,
		}

		metadata := map[string]string{
			"source": "event_consumer",
			"error":  err.Error(),
		}

		if dlqErr := ec.deadLetterQueue.AddEvent(ctx, userEvent.EventType, eventData, err, metadata); dlqErr != nil {
			ec.logger.Error("Failed to add event to dead letter queue: %v", dlqErr)
		} else {
			ec.logger.Warn("Event added to dead letter queue: %s, error: %v", userEvent.EventType, err)
		}

		return err
	}

	ec.logger.Info("Successfully processed event: %s for user: %s", userEvent.EventType, userEvent.UserID)
	return nil
}

// processEvent processes a single event
func (ec *EventConsumer) processEvent(ctx context.Context, event *entities.UserEvent) error {
	// Find and execute handler
	handler, exists := ec.eventHandlers[event.EventType]
	if !exists {
		return fmt.Errorf("no handler registered for event type: %s", event.EventType)
	}

	// Execute handler with retry logic
	return ec.executeWithRetry(ctx, func() error {
		return handler.HandleEvent(ctx, event)
	})
}

// executeWithRetry executes a function with retry logic
func (ec *EventConsumer) executeWithRetry(ctx context.Context, fn func() error) error {
	maxAttempts := 3
	delay := time.Second

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxAttempts {
				ec.logger.Warn("Attempt %d failed, retrying in %v: %v", attempt, delay, err)
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// GetDLQStats returns dead letter queue statistics
func (ec *EventConsumer) GetDLQStats(ctx context.Context) (resilience.DLQStats, error) {
	return ec.deadLetterQueue.GetStats(ctx)
}

// RetryFailedEvent retries a failed event from dead letter queue
func (ec *EventConsumer) RetryFailedEvent(ctx context.Context, eventID string) error {
	return ec.deadLetterQueue.RetryEvent(ctx, eventID)
}

// ListFailedEvents lists failed events from dead letter queue
func (ec *EventConsumer) ListFailedEvents(ctx context.Context, limit, offset int) ([]*resilience.FailedEvent, error) {
	return ec.deadLetterQueue.ListEvents(ctx, limit, offset)
}

// GetFailedEvent gets a specific failed event from dead letter queue
func (ec *EventConsumer) GetFailedEvent(ctx context.Context, eventID string) (*resilience.FailedEvent, error) {
	return ec.deadLetterQueue.GetEvent(ctx, eventID)
}

// DeleteFailedEvent deletes a failed event from dead letter queue
func (ec *EventConsumer) DeleteFailedEvent(ctx context.Context, eventID string) error {
	return ec.deadLetterQueue.DeleteEvent(ctx, eventID)
}
