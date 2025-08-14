package kafka

import (
	"github.com/IBM/sarama"

	"go-clean-ddd-es-template/pkg/metrics"
)

// ProducerWrapper wraps Kafka producer with metrics
type ProducerWrapper struct {
	producer sarama.SyncProducer
	metrics  *metrics.Metrics
}

// NewProducerWrapper creates a new Kafka producer wrapper
func NewProducerWrapper(producer sarama.SyncProducer, m *metrics.Metrics) *ProducerWrapper {
	return &ProducerWrapper{
		producer: producer,
		metrics:  m,
	}
}

// SendMessage wraps producer.SendMessage with metrics
func (w *ProducerWrapper) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	partition, offset, err := w.producer.SendMessage(msg)

	if err != nil {
		w.metrics.RecordKafkaProducerError(err.Error())
	} else {
		// Extract event type from message headers or key
		eventType := "unknown"
		if msg.Key != nil {
			if keyBytes, ok := msg.Key.(sarama.StringEncoder); ok {
				eventType = string(keyBytes)
			}
		}
		w.metrics.RecordKafkaEventPublished(msg.Topic, eventType)
	}

	return partition, offset, err
}

// SendMessages wraps producer.SendMessages with metrics
func (w *ProducerWrapper) SendMessages(msgs []*sarama.ProducerMessage) error {
	err := w.producer.SendMessages(msgs)

	if err != nil {
		w.metrics.RecordKafkaProducerError(err.Error())
	} else {
		// Record each message
		for _, msg := range msgs {
			eventType := "unknown"
			if msg.Key != nil {
				if keyBytes, ok := msg.Key.(sarama.StringEncoder); ok {
					eventType = string(keyBytes)
				}
			}
			w.metrics.RecordKafkaEventPublished(msg.Topic, eventType)
		}
	}

	return err
}

// Close wraps producer.Close
func (w *ProducerWrapper) Close() error {
	return w.producer.Close()
}

// ConsumerWrapper wraps Kafka consumer with metrics
type ConsumerWrapper struct {
	consumer sarama.Consumer
	metrics  *metrics.Metrics
}

// NewConsumerWrapper creates a new Kafka consumer wrapper
func NewConsumerWrapper(consumer sarama.Consumer, m *metrics.Metrics) *ConsumerWrapper {
	return &ConsumerWrapper{
		consumer: consumer,
		metrics:  m,
	}
}

// ConsumePartition wraps consumer.ConsumePartition with metrics
func (w *ConsumerWrapper) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	pc, err := w.consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		w.metrics.RecordKafkaProducerError(err.Error())
	}
	return pc, err
}

// Topics wraps consumer.Topics
func (w *ConsumerWrapper) Topics() ([]string, error) {
	return w.consumer.Topics()
}

// Partitions wraps consumer.Partitions
func (w *ConsumerWrapper) Partitions(topic string) ([]int32, error) {
	return w.consumer.Partitions(topic)
}

// GetConsumer returns the underlying consumer
func (w *ConsumerWrapper) GetConsumer() sarama.Consumer {
	return w.consumer
}

// Close wraps consumer.Close
func (w *ConsumerWrapper) Close() error {
	return w.consumer.Close()
}
