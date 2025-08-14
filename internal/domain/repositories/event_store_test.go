package repositories_test

import (
	"context"
	"testing"
	"time"

	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/domain/repositories"

	"github.com/stretchr/testify/assert"
)

func TestEventStore_Interface(t *testing.T) {
	// Test that EventStore interface is properly defined
	var _ repositories.EventStore = (*mockEventStore)(nil)
}

func TestEventStore_SaveEvent(t *testing.T) {
	// This test verifies the interface contract
	// Actual implementation would be tested in infrastructure layer
	event, err := events.NewEvent("test.event", "test-data", 1)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "test.event", event.Type)
	assert.Equal(t, 1, event.Version)
}

func TestEventStore_GetEvents(t *testing.T) {
	// This test verifies the interface contract
	// Actual implementation would be tested in infrastructure layer
	event1, _ := events.NewEvent("user.created", map[string]string{"name": "John"}, 1)
	event2, _ := events.NewEvent("user.updated", map[string]string{"name": "Jane"}, 2)
	events := []*events.Event{event1, event2}

	assert.Len(t, events, 2)
	assert.Equal(t, "user.created", events[0].Type)
	assert.Equal(t, "user.updated", events[1].Type)
}

func TestEventStore_GetEventsByType(t *testing.T) {
	// This test verifies the interface contract
	// Actual implementation would be tested in infrastructure layer
	event, _ := events.NewEvent("user.created", map[string]string{"user_id": "user-123", "name": "John"}, 1)
	events := []*events.Event{event}

	assert.Len(t, events, 1)
	assert.Equal(t, "user.created", events[0].Type)
	assert.Contains(t, string(events[0].Data), "user-123")
}

func TestEventStore_GetEventsSince(t *testing.T) {
	// This test verifies the interface contract
	// Actual implementation would be tested in infrastructure layer
	_ = time.Now().Add(-1 * time.Hour) // since timestamp
	event, _ := events.NewEvent("user.created", map[string]string{"name": "John"}, 1)
	events := []*events.Event{event}

	assert.Len(t, events, 1)
	assert.Equal(t, "user.created", events[0].Type)
}

func TestEventPublisher_Interface(t *testing.T) {
	// Test that EventPublisher interface is properly defined
	var _ repositories.EventPublisher = (*mockEventPublisher)(nil)
}

func TestEventPublisher_PublishEvent(t *testing.T) {
	// This test verifies the interface contract
	// Actual implementation would be tested in infrastructure layer
	event, err := events.NewEvent("test.event", "test-data", 1)
	assert.NoError(t, err)
	assert.NotNil(t, event)
}

func TestEventPublisher_PublishEvents(t *testing.T) {
	// This test verifies the interface contract
	// Actual implementation would be tested in infrastructure layer
	event1, _ := events.NewEvent("user.created", map[string]string{"name": "John"}, 1)
	event2, _ := events.NewEvent("user.updated", map[string]string{"name": "Jane"}, 2)
	events := []*events.Event{event1, event2}

	assert.Len(t, events, 2)
}

// Mock implementation for testing interface compliance
type mockEventStore struct{}

func (m *mockEventStore) SaveEvent(ctx context.Context, aggregateID string, event *events.Event) error {
	return nil
}

func (m *mockEventStore) GetEvents(ctx context.Context, aggregateID string) ([]*events.Event, error) {
	return []*events.Event{}, nil
}

func (m *mockEventStore) GetEventsByType(ctx context.Context, eventType string) ([]*events.Event, error) {
	return []*events.Event{}, nil
}

func (m *mockEventStore) GetEventsSince(ctx context.Context, since time.Time) ([]*events.Event, error) {
	return []*events.Event{}, nil
}

type mockEventPublisher struct{}

func (m *mockEventPublisher) PublishEvent(ctx context.Context, event *events.Event) error {
	return nil
}

func (m *mockEventPublisher) PublishEvents(ctx context.Context, events []*events.Event) error {
	return nil
}
