package repositories

import (
	"context"
	"time"

	"go-clean-ddd-es-template/internal/domain/events"
)

// EventStore defines the interface for event storage
type EventStore interface {
	// SaveEvent saves a domain event
	SaveEvent(ctx context.Context, aggregateID string, event *events.Event) error

	// GetEvents retrieves all events for a given aggregate ID
	GetEvents(ctx context.Context, aggregateID string) ([]*events.Event, error)

	// GetEventsByType retrieves events by type
	GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error)

	// GetEventsSince retrieves events since a given timestamp
	GetEventsSince(ctx context.Context, since time.Time) ([]*events.Event, error)
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// PublishEvent publishes a domain event
	PublishEvent(ctx context.Context, event *events.Event) error

	// PublishEvents publishes multiple domain events
	PublishEvents(ctx context.Context, events []*events.Event) error
}
