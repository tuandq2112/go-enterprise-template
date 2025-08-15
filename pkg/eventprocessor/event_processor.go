package eventprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Event represents a generic event
type Event interface {
	GetType() string
	GetData() map[string]interface{}
	GetTimestamp() time.Time
	GetVersion() int
	GetID() string
}

// EventHandler defines how events should be processed
type EventHandler interface {
	HandleEvent(ctx context.Context, event Event) error
	GetEventType() string
}

// EventProcessor handles event processing with multiple handlers
type EventProcessor struct {
	handlers map[string]EventHandler
	mu       sync.RWMutex
	logger   Logger
	metrics  *EventMetrics
}

// EventMetrics holds event processing metrics
type EventMetrics struct {
	mu              sync.RWMutex
	ProcessedEvents int64
	FailedEvents    int64
	RetryEvents     int64
	HandlerStats    map[string]*HandlerStats
}

// HandlerStats holds statistics for individual handlers
type HandlerStats struct {
	EventsProcessed int64
	EventsFailed    int64
	LastEventTime   time.Time
}

// Logger interface for logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// Config holds event processor configuration
type Config struct {
	MaxRetries    int           // Maximum number of retries per event
	RetryDelay    time.Duration // Delay between retries
	EnableMetrics bool          // Whether to enable metrics collection
}

// DefaultConfig returns default event processor configuration
func DefaultConfig() Config {
	return Config{
		MaxRetries:    3,
		RetryDelay:    time.Second,
		EnableMetrics: true,
	}
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(config Config, logger Logger) *EventProcessor {
	processor := &EventProcessor{
		handlers: make(map[string]EventHandler),
		logger:   logger,
		metrics:  &EventMetrics{HandlerStats: make(map[string]*HandlerStats)},
	}

	return processor
}

// RegisterHandler registers an event handler for a specific event type
func (ep *EventProcessor) RegisterHandler(handler EventHandler) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	eventType := handler.GetEventType()
	ep.handlers[eventType] = handler

	// Initialize handler stats
	if ep.metrics != nil {
		ep.metrics.mu.Lock()
		ep.metrics.HandlerStats[eventType] = &HandlerStats{}
		ep.metrics.mu.Unlock()
	}

	ep.logger.Info("Registered handler for event type: %s", eventType)
}

// UnregisterHandler unregisters an event handler
func (ep *EventProcessor) UnregisterHandler(eventType string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	delete(ep.handlers, eventType)
	ep.logger.Info("Unregistered handler for event type: %s", eventType)
}

// ProcessEvent processes a single event
func (ep *EventProcessor) ProcessEvent(ctx context.Context, event Event) error {
	ep.mu.RLock()
	handler, exists := ep.handlers[event.GetType()]
	ep.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for event type: %s", event.GetType())
	}

	// Process event with retry logic
	return ep.executeWithRetry(ctx, func() error {
		return handler.HandleEvent(ctx, event)
	}, event)
}

// ProcessEvents processes multiple events
func (ep *EventProcessor) ProcessEvents(ctx context.Context, events []Event) error {
	for _, event := range events {
		if err := ep.ProcessEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// ProcessRawEvent processes a raw event from message broker
func (ep *EventProcessor) ProcessRawEvent(ctx context.Context, rawEvent []byte, eventType string) error {
	// Parse raw event data
	var eventData map[string]interface{}
	if err := json.Unmarshal(rawEvent, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// Create generic event
	event := &GenericEvent{
		Type:      eventType,
		Data:      eventData,
		Timestamp: time.Now(),
		Version:   1,
		ID:        generateEventID(),
	}

	return ep.ProcessEvent(ctx, event)
}

// executeWithRetry executes a function with retry logic
func (ep *EventProcessor) executeWithRetry(ctx context.Context, fn func() error, event Event) error {
	maxAttempts := 3
	delay := time.Second

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err == nil {
			// Success - update metrics
			ep.updateMetrics(event.GetType(), true)
			return nil
		} else {
			lastErr = err
			if attempt < maxAttempts {
				ep.logger.Warn("Attempt %d failed for event %s, retrying in %v: %v",
					attempt, event.GetType(), delay, err)
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
		}
	}

	// All attempts failed - update metrics
	ep.updateMetrics(event.GetType(), false)
	return fmt.Errorf("failed to process event %s after %d attempts: %w",
		event.GetType(), maxAttempts, lastErr)
}

// updateMetrics updates event processing metrics
func (ep *EventProcessor) updateMetrics(eventType string, success bool) {
	if ep.metrics == nil {
		return
	}

	ep.metrics.mu.Lock()
	defer ep.metrics.mu.Unlock()

	if success {
		ep.metrics.ProcessedEvents++
		if stats, exists := ep.metrics.HandlerStats[eventType]; exists {
			stats.EventsProcessed++
			stats.LastEventTime = time.Now()
		}
	} else {
		ep.metrics.FailedEvents++
		if stats, exists := ep.metrics.HandlerStats[eventType]; exists {
			stats.EventsFailed++
		}
	}
}

// GetMetrics returns event processor metrics
func (ep *EventProcessor) GetMetrics() *EventMetrics {
	if ep.metrics == nil {
		return nil
	}

	ep.metrics.mu.RLock()
	defer ep.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &EventMetrics{
		ProcessedEvents: ep.metrics.ProcessedEvents,
		FailedEvents:    ep.metrics.FailedEvents,
		RetryEvents:     ep.metrics.RetryEvents,
		HandlerStats:    make(map[string]*HandlerStats),
	}

	for eventType, stats := range ep.metrics.HandlerStats {
		metrics.HandlerStats[eventType] = &HandlerStats{
			EventsProcessed: stats.EventsProcessed,
			EventsFailed:    stats.EventsFailed,
			LastEventTime:   stats.LastEventTime,
		}
	}

	return metrics
}

// GetRegisteredEventTypes returns all registered event types
func (ep *EventProcessor) GetRegisteredEventTypes() []string {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	eventTypes := make([]string, 0, len(ep.handlers))
	for eventType := range ep.handlers {
		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes
}

// HasHandler checks if a handler is registered for the given event type
func (ep *EventProcessor) HasHandler(eventType string) bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	_, exists := ep.handlers[eventType]
	return exists
}

// GetHandlerCount returns the number of registered handlers
func (ep *EventProcessor) GetHandlerCount() int {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	return len(ep.handlers)
}

// GenericEvent implements Event interface for generic event processing
type GenericEvent struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Version   int                    `json:"version"`
	ID        string                 `json:"id"`
}

func (e *GenericEvent) GetType() string {
	return e.Type
}

func (e *GenericEvent) GetData() map[string]interface{} {
	return e.Data
}

func (e *GenericEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *GenericEvent) GetVersion() int {
	return e.Version
}

func (e *GenericEvent) GetID() string {
	return e.ID
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}
