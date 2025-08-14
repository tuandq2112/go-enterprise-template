package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// EventHandler defines the interface for handling specific event types
type EventHandler interface {
	HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error
}

// EventConsumer is a generic consumer that can handle events for multiple models
type EventConsumer struct {
	consumer      sarama.Consumer
	eventHandlers map[string]EventHandler
	consumerGroup string
	topics        []string
}

// NewEventConsumer creates a new generic event consumer
func NewEventConsumer(
	consumer sarama.Consumer,
	consumerGroup string,
	topics []string,
) *EventConsumer {
	return &EventConsumer{
		consumer:      consumer,
		eventHandlers: make(map[string]EventHandler),
		consumerGroup: consumerGroup,
		topics:        topics,
	}
}

// RegisterEventHandler registers an event handler for a specific event type
func (c *EventConsumer) RegisterEventHandler(eventType string, handler EventHandler) {
	c.eventHandlers[eventType] = handler
}

// Start starts consuming events from Kafka
func (c *EventConsumer) Start(ctx context.Context) error {
	for _, topic := range c.topics {
		go c.consumeTopic(ctx, topic)
	}
	return nil
}

// consumeTopic consumes events from a specific topic
func (c *EventConsumer) consumeTopic(ctx context.Context, topic string) {
	partitions, err := c.consumer.Partitions(topic)
	if err != nil {
		log.Printf("Failed to get partitions for topic %s: %v", topic, err)
		return
	}

	for _, partition := range partitions {
		partitionConsumer, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Failed to create partition consumer for topic %s, partition %d: %v", topic, partition, err)
			continue
		}

		go c.consumePartition(ctx, partitionConsumer)
	}
}

// consumePartition consumes events from a specific partition
func (c *EventConsumer) consumePartition(ctx context.Context, partitionConsumer sarama.PartitionConsumer) {
	defer partitionConsumer.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-partitionConsumer.Messages():
			if err := c.handleMessage(ctx, msg); err != nil {
				log.Printf("Failed to handle message from topic %s: %v", msg.Topic, err)
			}
		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming from partition: %v", err)
		}
	}
}

// handleMessage handles a single Kafka message
func (c *EventConsumer) handleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))

	// Parse the event
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Extract event type
	eventType, ok := event["type"].(string)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	// Extract event data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data")
	}

	// Find and call appropriate handler
	handler, exists := c.eventHandlers[eventType]
	if !exists {
		log.Printf("No handler registered for event type: %s", eventType)
		return nil
	}

	// Handle the event
	if err := handler.HandleEvent(ctx, eventType, data); err != nil {
		return fmt.Errorf("failed to handle event %s: %w", eventType, err)
	}

	log.Printf("Successfully processed event: %s", eventType)
	return nil
}
