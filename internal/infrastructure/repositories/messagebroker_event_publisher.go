package repositories

import (
	"context"
	"encoding/json"

	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/infrastructure/messagebroker"
)

// MessageBrokerEventPublisher implements EventPublisher using message broker
type MessageBrokerEventPublisher struct {
	broker messagebroker.MessageBroker
}

// NewMessageBrokerEventPublisher creates a new message broker event publisher
func NewMessageBrokerEventPublisher(broker messagebroker.MessageBroker) *MessageBrokerEventPublisher {
	return &MessageBrokerEventPublisher{
		broker: broker,
	}
}

// PublishEvent publishes an event to the message broker
func (p *MessageBrokerEventPublisher) PublishEvent(ctx context.Context, event *events.Event) error {
	// Serialize event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish to message broker
	topic := event.Type // Use event type as topic
	return p.broker.Publish(topic, eventData)
}

// PublishEvents publishes multiple events to the message broker
func (p *MessageBrokerEventPublisher) PublishEvents(ctx context.Context, events []*events.Event) error {
	for _, event := range events {
		if err := p.PublishEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
