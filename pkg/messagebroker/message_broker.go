package messagebroker

import (
	"context"
	"time"
)

// Message represents a generic message
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string][]byte
	Timestamp time.Time
	Partition int32
	Offset    int64
}

// MessageHandler defines how messages should be handled
type MessageHandler func(ctx context.Context, message *Message) error

// MessageBroker defines the interface for message brokers
type MessageBroker interface {
	// Connection management
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	IsConnected() bool

	// Publishing
	Publish(ctx context.Context, topic string, message *Message) error
	PublishBatch(ctx context.Context, topic string, messages []*Message) error

	// Consuming
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Unsubscribe(ctx context.Context, topic string) error

	// Topic management
	CreateTopic(ctx context.Context, topic string, partitions int) error
	DeleteTopic(ctx context.Context, topic string) error
	TopicExists(ctx context.Context, topic string) (bool, error)

	// Consumer group management
	CreateConsumerGroup(ctx context.Context, groupID string) error
	JoinConsumerGroup(ctx context.Context, groupID string, topics []string) error
	LeaveConsumerGroup(ctx context.Context, groupID string) error

	// Health and monitoring
	Health(ctx context.Context) error
	GetStats(ctx context.Context) (*BrokerStats, error)
}

// BrokerStats holds message broker statistics
type BrokerStats struct {
	PublishedMessages int64
	ConsumedMessages  int64
	FailedMessages    int64
	ActiveTopics      int
	ActiveConsumers   int
	ConnectionStatus  string
	LastActivity      time.Time
}

// Config holds message broker configuration
type Config struct {
	// Connection settings
	Brokers  []string
	ClientID string
	GroupID  string

	// Timeout settings
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration

	// Retry settings
	MaxRetries int
	RetryDelay time.Duration

	// Security settings
	Username    string
	Password    string
	SSLEnabled  bool
	SSLCertFile string
	SSLKeyFile  string
	SSLCAFile   string

	// Producer settings
	ProducerAcks    string // "none", "leader", "all"
	ProducerRetries int

	// Consumer settings
	ConsumerAutoCommit bool
	ConsumerOffset     string // "earliest", "latest"

	// Worker pool settings
	PublisherWorkers int
	ConsumerWorkers  int
	WorkerBufferSize int
}

// DefaultConfig returns default message broker configuration
func DefaultConfig() Config {
	return Config{
		Brokers:            []string{"localhost:9092"},
		ClientID:           "default-client",
		GroupID:            "default-group",
		ConnectTimeout:     30 * time.Second,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		MaxRetries:         3,
		RetryDelay:         time.Second,
		SSLEnabled:         false,
		ProducerAcks:       "all",
		ProducerRetries:    3,
		ConsumerAutoCommit: true,
		ConsumerOffset:     "latest",
		PublisherWorkers:   10,
		ConsumerWorkers:    20,
		WorkerBufferSize:   1000,
	}
}

// MessageBuilder helps build messages
type MessageBuilder struct {
	message *Message
}

// NewMessageBuilder creates a new message builder
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		message: &Message{
			Headers:   make(map[string][]byte),
			Timestamp: time.Now(),
		},
	}
}

// WithTopic sets the message topic
func (mb *MessageBuilder) WithTopic(topic string) *MessageBuilder {
	mb.message.Topic = topic
	return mb
}

// WithKey sets the message key
func (mb *MessageBuilder) WithKey(key []byte) *MessageBuilder {
	mb.message.Key = key
	return mb
}

// WithValue sets the message value
func (mb *MessageBuilder) WithValue(value []byte) *MessageBuilder {
	mb.message.Value = value
	return mb
}

// WithHeader adds a header to the message
func (mb *MessageBuilder) WithHeader(key string, value []byte) *MessageBuilder {
	mb.message.Headers[key] = value
	return mb
}

// WithHeaders sets multiple headers
func (mb *MessageBuilder) WithHeaders(headers map[string][]byte) *MessageBuilder {
	for k, v := range headers {
		mb.message.Headers[k] = v
	}
	return mb
}

// WithTimestamp sets the message timestamp
func (mb *MessageBuilder) WithTimestamp(timestamp time.Time) *MessageBuilder {
	mb.message.Timestamp = timestamp
	return mb
}

// Build builds the final message
func (mb *MessageBuilder) Build() *Message {
	return mb.message
}

// ConsumerGroup represents a consumer group
type ConsumerGroup struct {
	GroupID string
	Topics  []string
	Handler MessageHandler
}

// NewConsumerGroup creates a new consumer group
func NewConsumerGroup(groupID string, topics []string, handler MessageHandler) *ConsumerGroup {
	return &ConsumerGroup{
		GroupID: groupID,
		Topics:  topics,
		Handler: handler,
	}
}

// TopicConfig holds topic configuration
type TopicConfig struct {
	Name       string
	Partitions int
	Replicas   int
	Config     map[string]string
}

// NewTopicConfig creates a new topic configuration
func NewTopicConfig(name string, partitions int) *TopicConfig {
	return &TopicConfig{
		Name:       name,
		Partitions: partitions,
		Replicas:   1,
		Config:     make(map[string]string),
	}
}

// WithReplicas sets the number of replicas
func (tc *TopicConfig) WithReplicas(replicas int) *TopicConfig {
	tc.Replicas = replicas
	return tc
}

// WithConfig sets a topic configuration parameter
func (tc *TopicConfig) WithConfig(key, value string) *TopicConfig {
	tc.Config[key] = value
	return tc
}
