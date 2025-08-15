package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FailedEvent represents a failed event in the dead letter queue
type FailedEvent struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	EventData   map[string]interface{} `json:"event_data"`
	Error       string                 `json:"error"`
	Timestamp   time.Time              `json:"timestamp"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Topic       string                 `json:"topic"`
	Partition   int32                  `json:"partition"`
	Offset      int64                  `json:"offset"`
	Metadata    map[string]string      `json:"metadata"`
}

// DeadLetterQueue manages failed events
type DeadLetterQueue struct {
	mu sync.RWMutex

	// Configuration
	maxSize      int
	maxAttempts  int
	retryDelay   time.Duration
	storage      DLQStorage
	retryHandler RetryHandler

	// In-memory storage (fallback)
	events []*FailedEvent
}

// DLQStorage interface for persistent storage
type DLQStorage interface {
	Store(ctx context.Context, event *FailedEvent) error
	Get(ctx context.Context, id string) (*FailedEvent, error)
	List(ctx context.Context, limit, offset int) ([]*FailedEvent, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
}

// RetryHandler interface for retry logic
type RetryHandler interface {
	HandleRetry(ctx context.Context, event *FailedEvent) error
}

// DeadLetterQueueConfig holds configuration for DLQ
type DeadLetterQueueConfig struct {
	MaxSize     int           `json:"max_size"`
	MaxAttempts int           `json:"max_attempts"`
	RetryDelay  time.Duration `json:"retry_delay"`
}

// DefaultDeadLetterQueueConfig returns default configuration
func DefaultDeadLetterQueueConfig() DeadLetterQueueConfig {
	return DeadLetterQueueConfig{
		MaxSize:     1000,
		MaxAttempts: 3,
		RetryDelay:  5 * time.Minute,
	}
}

// NewDeadLetterQueue creates a new dead letter queue
func NewDeadLetterQueue(config DeadLetterQueueConfig, storage DLQStorage, retryHandler RetryHandler) *DeadLetterQueue {
	return &DeadLetterQueue{
		maxSize:      config.MaxSize,
		maxAttempts:  config.MaxAttempts,
		retryDelay:   config.RetryDelay,
		storage:      storage,
		retryHandler: retryHandler,
		events:       make([]*FailedEvent, 0),
	}
}

// AddEvent adds a failed event to the dead letter queue
func (dlq *DeadLetterQueue) AddEvent(ctx context.Context, eventType string, eventData map[string]interface{}, err error, metadata map[string]string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	// Check if queue is full
	if len(dlq.events) >= dlq.maxSize {
		return fmt.Errorf("dead letter queue is full (max size: %d)", dlq.maxSize)
	}

	failedEvent := &FailedEvent{
		ID:          generateEventID(),
		EventType:   eventType,
		EventData:   eventData,
		Error:       err.Error(),
		Timestamp:   time.Now(),
		Attempts:    0,
		MaxAttempts: dlq.maxAttempts,
		Metadata:    metadata,
	}

	// Try to store in persistent storage first
	if dlq.storage != nil {
		if storeErr := dlq.storage.Store(ctx, failedEvent); storeErr != nil {
			// Fallback to in-memory storage
			dlq.events = append(dlq.events, failedEvent)
			return fmt.Errorf("failed to store in persistent storage: %w, stored in memory", storeErr)
		}
	} else {
		// Use in-memory storage
		dlq.events = append(dlq.events, failedEvent)
	}

	return nil
}

// AddKafkaEvent adds a failed Kafka event to the dead letter queue
func (dlq *DeadLetterQueue) AddKafkaEvent(ctx context.Context, eventType string, eventData map[string]interface{}, err error, topic string, partition int32, offset int64) error {
	metadata := map[string]string{
		"source": "kafka",
		"topic":  topic,
	}

	failedEvent := &FailedEvent{
		ID:          generateEventID(),
		EventType:   eventType,
		EventData:   eventData,
		Error:       err.Error(),
		Timestamp:   time.Now(),
		Attempts:    0,
		MaxAttempts: dlq.maxAttempts,
		Topic:       topic,
		Partition:   partition,
		Offset:      offset,
		Metadata:    metadata,
	}

	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	// Check if queue is full
	if len(dlq.events) >= dlq.maxSize {
		return fmt.Errorf("dead letter queue is full (max size: %d)", dlq.maxSize)
	}

	// Try to store in persistent storage first
	if dlq.storage != nil {
		if storeErr := dlq.storage.Store(ctx, failedEvent); storeErr != nil {
			// Fallback to in-memory storage
			dlq.events = append(dlq.events, failedEvent)
			return fmt.Errorf("failed to store in persistent storage: %w, stored in memory", storeErr)
		}
	} else {
		// Use in-memory storage
		dlq.events = append(dlq.events, failedEvent)
	}

	return nil
}

// RetryEvent attempts to retry a failed event
func (dlq *DeadLetterQueue) RetryEvent(ctx context.Context, eventID string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	var event *FailedEvent
	var err error

	// Try to get from persistent storage first
	if dlq.storage != nil {
		event, err = dlq.storage.Get(ctx, eventID)
		if err != nil {
			return fmt.Errorf("failed to get event from storage: %w", err)
		}
	} else {
		// Find in memory
		event = dlq.findEventByID(eventID)
		if event == nil {
			return fmt.Errorf("event not found: %s", eventID)
		}
	}

	// Check if max attempts reached
	if event.Attempts >= event.MaxAttempts {
		return fmt.Errorf("max retry attempts reached for event %s", eventID)
	}

	// Increment attempts
	event.Attempts++

	// Try to retry
	if dlq.retryHandler != nil {
		if retryErr := dlq.retryHandler.HandleRetry(ctx, event); retryErr != nil {
			// Update error message
			event.Error = retryErr.Error()
			event.Timestamp = time.Now()

			// Update in storage
			if dlq.storage != nil {
				if updateErr := dlq.storage.Store(ctx, event); updateErr != nil {
					return fmt.Errorf("failed to update event in storage: %w", updateErr)
				}
			}

			return fmt.Errorf("retry failed: %w", retryErr)
		}
	}

	// Success - remove from queue
	if dlq.storage != nil {
		if deleteErr := dlq.storage.Delete(ctx, eventID); deleteErr != nil {
			return fmt.Errorf("failed to delete event from storage: %w", deleteErr)
		}
	} else {
		dlq.removeEventByID(eventID)
	}

	return nil
}

// GetEvent retrieves a failed event by ID
func (dlq *DeadLetterQueue) GetEvent(ctx context.Context, eventID string) (*FailedEvent, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	if dlq.storage != nil {
		return dlq.storage.Get(ctx, eventID)
	}

	event := dlq.findEventByID(eventID)
	if event == nil {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	return event, nil
}

// ListEvents lists failed events with pagination
func (dlq *DeadLetterQueue) ListEvents(ctx context.Context, limit, offset int) ([]*FailedEvent, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	if dlq.storage != nil {
		return dlq.storage.List(ctx, limit, offset)
	}

	// In-memory pagination
	if offset >= len(dlq.events) {
		return []*FailedEvent{}, nil
	}

	end := offset + limit
	if end > len(dlq.events) {
		end = len(dlq.events)
	}

	// Create a copy to avoid race conditions
	events := make([]*FailedEvent, end-offset)
	copy(events, dlq.events[offset:end])

	return events, nil
}

// DeleteEvent removes a failed event from the queue
func (dlq *DeadLetterQueue) DeleteEvent(ctx context.Context, eventID string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if dlq.storage != nil {
		return dlq.storage.Delete(ctx, eventID)
	}

	if dlq.removeEventByID(eventID) {
		return nil
	}

	return fmt.Errorf("event not found: %s", eventID)
}

// GetStats returns dead letter queue statistics
func (dlq *DeadLetterQueue) GetStats(ctx context.Context) (DLQStats, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	var count int
	var err error

	if dlq.storage != nil {
		count, err = dlq.storage.Count(ctx)
		if err != nil {
			return DLQStats{}, fmt.Errorf("failed to get count from storage: %w", err)
		}
	} else {
		count = len(dlq.events)
	}

	return DLQStats{
		TotalEvents: count,
		MaxSize:     dlq.maxSize,
		MaxAttempts: dlq.maxAttempts,
		RetryDelay:  dlq.retryDelay,
		Utilization: float64(count) / float64(dlq.maxSize) * 100,
	}, nil
}

// Clear removes all events from the queue
func (dlq *DeadLetterQueue) Clear(ctx context.Context) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if dlq.storage != nil {
		// For persistent storage, we need to implement a clear method
		// For now, we'll just clear the in-memory events
		dlq.events = make([]*FailedEvent, 0)
		return nil
	}

	dlq.events = make([]*FailedEvent, 0)
	return nil
}

// DLQStats holds statistics for dead letter queue
type DLQStats struct {
	TotalEvents int           `json:"total_events"`
	MaxSize     int           `json:"max_size"`
	MaxAttempts int           `json:"max_attempts"`
	RetryDelay  time.Duration `json:"retry_delay"`
	Utilization float64       `json:"utilization_percent"`
}

// Helper methods for in-memory storage
func (dlq *DeadLetterQueue) findEventByID(eventID string) *FailedEvent {
	for _, event := range dlq.events {
		if event.ID == eventID {
			return event
		}
	}
	return nil
}

func (dlq *DeadLetterQueue) removeEventByID(eventID string) bool {
	for i, event := range dlq.events {
		if event.ID == eventID {
			dlq.events = append(dlq.events[:i], dlq.events[i+1:]...)
			return true
		}
	}
	return false
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("dlq_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
