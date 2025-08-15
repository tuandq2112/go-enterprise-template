package repositories

import (
	"context"
	"encoding/json"

	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/internal/infrastructure/messagebroker"
)

// MessageBrokerEventPublisher implements EventPublisher using message broker
type MessageBrokerEventPublisher struct {
	broker messagebroker.MessageBroker
	config *config.Config
}

// NewMessageBrokerEventPublisher creates a new message broker event publisher
func NewMessageBrokerEventPublisher(broker messagebroker.MessageBroker, config *config.Config) *MessageBrokerEventPublisher {
	return &MessageBrokerEventPublisher{
		broker: broker,
		config: config,
	}
}

// PublishEvent publishes an event to the message broker
func (p *MessageBrokerEventPublisher) PublishEvent(ctx context.Context, event *events.Event) error {
	// Serialize event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Get topic from config mapping, fallback to event type
	topic := p.getTopicForEvent(event.Type)
	return p.broker.Publish(topic, eventData)
}

// getTopicForEvent returns the appropriate topic for an event type
func (p *MessageBrokerEventPublisher) getTopicForEvent(eventType string) string {
	// Check if there's a mapping in config
	if mappedTopic, exists := p.config.MessageBroker.Topics[eventType]; exists {
		return mappedTopic
	}

	// Fallback to event type as topic name
	return eventType
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
