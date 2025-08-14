package messagebroker

import (
	"fmt"
	"log"

	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/pkg/kafka"
	"go-clean-ddd-es-template/pkg/metrics"

	"github.com/IBM/sarama"
)

// MessageBroker interface for different message broker types
type MessageBroker interface {
	Connect() error
	Close() error
	Publish(topic string, message []byte) error
	Subscribe(topic string, handler func([]byte)) error
	GetConsumer() sarama.Consumer
}

// MessageBrokerFactory creates message broker instances based on configuration
type MessageBrokerFactory struct{}

// NewMessageBrokerFactory creates a new message broker factory
func NewMessageBrokerFactory() *MessageBrokerFactory {
	return &MessageBrokerFactory{}
}

// CreateMessageBroker creates a message broker instance based on configuration
func (f *MessageBrokerFactory) CreateMessageBroker(cfg *config.MessageBrokerConfig) (MessageBroker, error) {
	switch cfg.Type {
	case "kafka":
		return NewKafkaBroker(cfg)
	case "rabbitmq":
		return NewRabbitMQBroker(cfg)
	case "redis":
		return NewRedisBroker(cfg)
	case "nats":
		return NewNATSBroker(cfg)
	default:
		return nil, fmt.Errorf("unsupported message broker type: %s", cfg.Type)
	}
}

// KafkaBroker implements MessageBroker interface using Kafka
type KafkaBroker struct {
	config   *config.MessageBrokerConfig
	producer *kafka.ProducerWrapper
	consumer *kafka.ConsumerWrapper
	metrics  *metrics.Metrics
}

func NewKafkaBroker(cfg *config.MessageBrokerConfig) (*KafkaBroker, error) {
	// Create Sarama config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	// Create Sarama producer
	saramaProducer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Create Sarama consumer
	saramaConsumer, err := sarama.NewConsumer(cfg.Brokers, nil)
	if err != nil {
		saramaProducer.Close()
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	// Create metrics instance
	m := metrics.NewMetrics()

	// Create wrapped producer and consumer
	producer := kafka.NewProducerWrapper(saramaProducer, m)
	consumer := kafka.NewConsumerWrapper(saramaConsumer, m)

	return &KafkaBroker{
		config:   cfg,
		producer: producer,
		consumer: consumer,
		metrics:  m,
	}, nil
}

func (k *KafkaBroker) Connect() error {
	// Connection is established in constructor
	log.Printf("Connected to Kafka brokers: %v", k.config.Brokers)
	return nil
}

func (k *KafkaBroker) Close() error {
	var errs []error

	if err := k.producer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close producer: %w", err))
	}

	if err := k.consumer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close consumer: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Kafka broker: %v", errs)
	}

	return nil
}

func (k *KafkaBroker) Publish(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}

	log.Printf("Message published to topic: %s", topic)
	return nil
}

func (k *KafkaBroker) Subscribe(topic string, handler func([]byte)) error {
	// Get partitions for the topic
	partitions, err := k.consumer.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
	}

	// Subscribe to all partitions
	for _, partition := range partitions {
		partitionConsumer, err := k.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to create partition consumer for topic %s, partition %d: %w", topic, partition, err)
		}

		// Start consuming in a goroutine
		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for msg := range pc.Messages() {
				handler(msg.Value)
			}
		}(partitionConsumer)
	}

	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

func (k *KafkaBroker) GetConsumer() sarama.Consumer {
	return k.consumer.GetConsumer()
}

// RabbitMQBroker stub implementation
type RabbitMQBroker struct {
	config *config.MessageBrokerConfig
}

func NewRabbitMQBroker(cfg *config.MessageBrokerConfig) (*RabbitMQBroker, error) {
	broker := &RabbitMQBroker{
		config: cfg,
	}

	if err := broker.Connect(); err != nil {
		return nil, err
	}

	return broker, nil
}

func (r *RabbitMQBroker) Connect() error {
	return fmt.Errorf("RabbitMQ implementation not available - use Kafka instead")
}

func (r *RabbitMQBroker) Close() error {
	return nil
}

func (r *RabbitMQBroker) Publish(topic string, message []byte) error {
	return fmt.Errorf("RabbitMQ implementation not available")
}

func (r *RabbitMQBroker) Subscribe(topic string, handler func([]byte)) error {
	return fmt.Errorf("RabbitMQ implementation not available")
}

func (r *RabbitMQBroker) GetConsumer() sarama.Consumer {
	return nil
}

// RedisBroker stub implementation
type RedisBroker struct {
	config *config.MessageBrokerConfig
}

func NewRedisBroker(cfg *config.MessageBrokerConfig) (*RedisBroker, error) {
	broker := &RedisBroker{
		config: cfg,
	}

	if err := broker.Connect(); err != nil {
		return nil, err
	}

	return broker, nil
}

func (r *RedisBroker) Connect() error {
	return fmt.Errorf("Redis implementation not available - use Kafka instead")
}

func (r *RedisBroker) Close() error {
	return nil
}

func (r *RedisBroker) Publish(topic string, message []byte) error {
	return fmt.Errorf("Redis implementation not available")
}

func (r *RedisBroker) Subscribe(topic string, handler func([]byte)) error {
	return fmt.Errorf("Redis implementation not available")
}

func (r *RedisBroker) GetConsumer() sarama.Consumer {
	return nil
}

// NATSBroker stub implementation
type NATSBroker struct {
	config *config.MessageBrokerConfig
}

func NewNATSBroker(cfg *config.MessageBrokerConfig) (*NATSBroker, error) {
	broker := &NATSBroker{
		config: cfg,
	}

	if err := broker.Connect(); err != nil {
		return nil, err
	}

	return broker, nil
}

func (n *NATSBroker) Connect() error {
	return fmt.Errorf("NATS implementation not available - use Kafka instead")
}

func (n *NATSBroker) Close() error {
	return nil
}

func (n *NATSBroker) Publish(topic string, message []byte) error {
	return fmt.Errorf("NATS implementation not available")
}

func (n *NATSBroker) Subscribe(topic string, handler func([]byte)) error {
	return fmt.Errorf("NATS implementation not available")
}

func (n *NATSBroker) GetConsumer() sarama.Consumer {
	return nil
}
