package consumers

import (
	"context"
	"log"
	"sync"

	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/infrastructure/config"

	"github.com/IBM/sarama"
)

// LegacyEventHandler represents the old event handler interface
type LegacyEventHandler interface {
	HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error
}

// EventHandlerAdapter adapts LegacyEventHandler to new EventHandler interface
type EventHandlerAdapter struct {
	legacyHandler LegacyEventHandler
}

// NewEventHandlerAdapter creates a new event handler adapter
func NewEventHandlerAdapter(legacyHandler LegacyEventHandler) *EventHandlerAdapter {
	return &EventHandlerAdapter{
		legacyHandler: legacyHandler,
	}
}

// HandleEvent adapts the new interface to the old one
func (a *EventHandlerAdapter) HandleEvent(ctx context.Context, event *entities.UserEvent) error {
	// Convert UserEvent to the old format
	eventData := make(map[string]interface{})
	if event.EventData != nil {
		eventData = event.EventData
	}

	return a.legacyHandler.HandleEvent(ctx, event.EventType, eventData)
}

// EventConsumerInterface defines the common interface for event consumers
type EventConsumerInterface interface {
	RegisterHandler(eventType string, handler EventHandler)
	HandleMessage(ctx context.Context, message []byte) error
}

// EventConsumerWrapper wraps the new EventConsumer to maintain compatibility
type EventConsumerWrapper struct {
	consumer      sarama.Consumer
	eventConsumer EventConsumerInterface
	consumerGroup string
	topics        []string
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewEventConsumerWrapper creates a new event consumer wrapper
func NewEventConsumerWrapper(consumer sarama.Consumer, consumerGroup string, topics []string) *EventConsumerWrapper {
	// Create a simple logger
	logger := &SimpleLogger{}

	// Create event consumer with default config
	config := DefaultEventConsumerConfig()
	eventConsumer := NewEventConsumer(config, logger)

	return &EventConsumerWrapper{
		consumer:      consumer,
		eventConsumer: eventConsumer,
		consumerGroup: consumerGroup,
		topics:        topics,
		stopChan:      make(chan struct{}),
	}
}

// NewEventConsumerWrapperWithWorkerPool creates a new event consumer wrapper with worker pool
func NewEventConsumerWrapperWithWorkerPool(consumer sarama.Consumer, consumerGroup string, topics []string, config *config.Config, logger Logger) *EventConsumerWrapper {
	// Create worker pool event consumer
	eventConsumer := NewWorkerPoolEventConsumer(config, consumer, logger)

	return &EventConsumerWrapper{
		consumer:      consumer,
		eventConsumer: eventConsumer,
		consumerGroup: consumerGroup,
		topics:        topics,
		stopChan:      make(chan struct{}),
	}
}

// RegisterEventHandler registers an event handler (compatibility method)
func (w *EventConsumerWrapper) RegisterEventHandler(eventType string, handler LegacyEventHandler) {
	// Create adapter for the legacy handler
	adapter := NewEventHandlerAdapter(handler)
	w.eventConsumer.RegisterHandler(eventType, adapter)
}

// Start starts the event consumer (compatibility method)
func (w *EventConsumerWrapper) Start(ctx context.Context) error {
	log.Printf("Starting event consumer for topics: %v", w.topics)

	// Start consuming from each topic
	for _, topic := range w.topics {
		w.wg.Add(1)
		go w.consumeTopic(ctx, topic)
	}

	log.Printf("Event consumer started successfully")
	return nil
}

// consumeTopic consumes messages from a specific topic
func (w *EventConsumerWrapper) consumeTopic(ctx context.Context, topic string) {
	defer w.wg.Done()

	// Get partition list for the topic
	partitions, err := w.consumer.Partitions(topic)
	if err != nil {
		log.Printf("[ERROR] Failed to get partitions for topic %s: %v", topic, err)
		return
	}

	// Create partition consumers
	for _, partition := range partitions {
		partitionConsumer, err := w.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Printf("[ERROR] Failed to create partition consumer for topic %s partition %d: %v", topic, partition, err)
			continue
		}
		defer partitionConsumer.Close()

		// Consume messages
		for {
			select {
			case <-ctx.Done():
				log.Printf("[INFO] Context cancelled, stopping consumer for topic %s partition %d", topic, partition)
				return
			case <-w.stopChan:
				log.Printf("[INFO] Stop signal received, stopping consumer for topic %s partition %d", topic, partition)
				return
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					log.Printf("[INFO] Received message from topic %s partition %d offset %d", topic, partition, msg.Offset)

					// Handle the message
					if err := w.eventConsumer.HandleMessage(ctx, msg.Value); err != nil {
						log.Printf("[ERROR] Failed to handle message from topic %s: %v", topic, err)
					}
				}
			case err := <-partitionConsumer.Errors():
				if err != nil {
					log.Printf("[ERROR] Error consuming from topic %s partition %d: %v", topic, partition, err)
				}
			}
		}
	}
}

// Stop stops the event consumer
func (w *EventConsumerWrapper) Stop() {
	log.Printf("[INFO] Stopping event consumer...")
	close(w.stopChan)
	w.wg.Wait()
	log.Printf("[INFO] Event consumer stopped")
}

// SimpleLogger implements the Logger interface
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}
