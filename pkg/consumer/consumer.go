package consumer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Message represents a message from the broker
type Message struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string][]byte
	Timestamp time.Time
}

// MessageHandler defines how messages should be processed
type MessageHandler func(ctx context.Context, message *Message) error

// Consumer represents a message consumer
type Consumer interface {
	// Basic operations
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Message processing
	Subscribe(topic string, handler MessageHandler) error
	Unsubscribe(topic string) error

	// Consumer group operations
	JoinGroup(groupID string) error
	LeaveGroup() error

	// Health and monitoring
	Health(ctx context.Context) error
	GetStats(ctx context.Context) (*ConsumerStats, error)
}

// ConsumerGroup represents a consumer group
type ConsumerGroup struct {
	GroupID   string
	Topics    []string
	Consumers []Consumer
	Handler   MessageHandler
	Config    *GroupConfig
}

// GroupConfig holds consumer group configuration
type GroupConfig struct {
	GroupID            string
	Topics             []string
	Handler            MessageHandler
	NumConsumers       int
	WorkerPoolSize     int
	AutoCommit         bool
	AutoCommitInterval time.Duration
	SessionTimeout     time.Duration
	HeartbeatInterval  time.Duration
	MaxPollRecords     int
	MaxPollInterval    time.Duration
}

// DefaultGroupConfig returns default consumer group configuration
func DefaultGroupConfig() *GroupConfig {
	return &GroupConfig{
		NumConsumers:       1,
		WorkerPoolSize:     10,
		AutoCommit:         true,
		AutoCommitInterval: 5 * time.Second,
		SessionTimeout:     30 * time.Second,
		HeartbeatInterval:  3 * time.Second,
		MaxPollRecords:     500,
		MaxPollInterval:    300 * time.Second,
	}
}

// ConsumerStats holds consumer statistics
type ConsumerStats struct {
	mu               sync.RWMutex
	GroupID          string
	Topics           []string
	MessagesConsumed int64
	MessagesFailed   int64
	MessagesRetried  int64
	LastMessageTime  time.Time
	ConsumerLag      map[string]int64 // topic -> lag
	ActiveConsumers  int
	IsRunning        bool
}

// ConsumerManager manages multiple consumers
type ConsumerManager struct {
	consumers map[string]Consumer
	groups    map[string]*ConsumerGroup
	mu        sync.RWMutex
	config    *ManagerConfig
	stats     *ManagerStats
}

// ManagerConfig holds consumer manager configuration
type ManagerConfig struct {
	MaxConsumers        int
	MaxGroups           int
	HealthCheckInterval time.Duration
	EnableMetrics       bool
}

// DefaultManagerConfig returns default consumer manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		MaxConsumers:        10,
		MaxGroups:           5,
		HealthCheckInterval: 30 * time.Second,
		EnableMetrics:       true,
	}
}

// ManagerStats holds consumer manager statistics
type ManagerStats struct {
	mu              sync.RWMutex
	TotalConsumers  int
	TotalGroups     int
	ActiveConsumers int
	ActiveGroups    int
	TotalMessages   int64
	FailedMessages  int64
}

// NewConsumerManager creates a new consumer manager
func NewConsumerManager(config *ManagerConfig) *ConsumerManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	return &ConsumerManager{
		consumers: make(map[string]Consumer),
		groups:    make(map[string]*ConsumerGroup),
		config:    config,
		stats:     &ManagerStats{},
	}
}

// CreateConsumer creates a new consumer
func (cm *ConsumerManager) CreateConsumer(consumerID string, consumer Consumer) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.consumers) >= cm.config.MaxConsumers {
		return fmt.Errorf("maximum number of consumers reached: %d", cm.config.MaxConsumers)
	}

	cm.consumers[consumerID] = consumer
	cm.stats.mu.Lock()
	cm.stats.TotalConsumers++
	cm.stats.mu.Unlock()

	log.Printf("Created consumer: %s", consumerID)
	return nil
}

// RemoveConsumer removes a consumer
func (cm *ConsumerManager) RemoveConsumer(consumerID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	consumer, exists := cm.consumers[consumerID]
	if !exists {
		return fmt.Errorf("consumer not found: %s", consumerID)
	}

	// Stop consumer
	ctx := context.Background()
	if err := consumer.Stop(ctx); err != nil {
		log.Printf("Error stopping consumer %s: %v", consumerID, err)
	}

	delete(cm.consumers, consumerID)
	cm.stats.mu.Lock()
	cm.stats.TotalConsumers--
	cm.stats.mu.Unlock()

	log.Printf("Removed consumer: %s", consumerID)
	return nil
}

// CreateConsumerGroup creates a new consumer group
func (cm *ConsumerManager) CreateConsumerGroup(config *GroupConfig) (*ConsumerGroup, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.groups) >= cm.config.MaxGroups {
		return nil, fmt.Errorf("maximum number of consumer groups reached: %d", cm.config.MaxGroups)
	}

	group := &ConsumerGroup{
		GroupID:   config.GroupID,
		Topics:    config.Topics,
		Handler:   config.Handler,
		Config:    config,
		Consumers: make([]Consumer, 0, config.NumConsumers),
	}

	cm.groups[config.GroupID] = group
	cm.stats.mu.Lock()
	cm.stats.TotalGroups++
	cm.stats.mu.Unlock()

	log.Printf("Created consumer group: %s", config.GroupID)
	return group, nil
}

// RemoveConsumerGroup removes a consumer group
func (cm *ConsumerManager) RemoveConsumerGroup(groupID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	group, exists := cm.groups[groupID]
	if !exists {
		return fmt.Errorf("consumer group not found: %s", groupID)
	}

	// Stop all consumers in the group
	ctx := context.Background()
	for _, consumer := range group.Consumers {
		if err := consumer.Stop(ctx); err != nil {
			log.Printf("Error stopping consumer in group %s: %v", groupID, err)
		}
	}

	delete(cm.groups, groupID)
	cm.stats.mu.Lock()
	cm.stats.TotalGroups--
	cm.stats.mu.Unlock()

	log.Printf("Removed consumer group: %s", groupID)
	return nil
}

// StartAll starts all consumers
func (cm *ConsumerManager) StartAll(ctx context.Context) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var errs []error

	// Start individual consumers
	for consumerID, consumer := range cm.consumers {
		if err := consumer.Start(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to start consumer %s: %w", consumerID, err))
		} else {
			cm.stats.mu.Lock()
			cm.stats.ActiveConsumers++
			cm.stats.mu.Unlock()
		}
	}

	// Start consumer groups
	for groupID, group := range cm.groups {
		if err := cm.startConsumerGroup(ctx, group); err != nil {
			errs = append(errs, fmt.Errorf("failed to start consumer group %s: %w", groupID, err))
		} else {
			cm.stats.mu.Lock()
			cm.stats.ActiveGroups++
			cm.stats.mu.Unlock()
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors starting consumers: %v", errs)
	}

	return nil
}

// StopAll stops all consumers
func (cm *ConsumerManager) StopAll(ctx context.Context) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var errs []error

	// Stop individual consumers
	for consumerID, consumer := range cm.consumers {
		if err := consumer.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop consumer %s: %w", consumerID, err))
		}
	}

	// Stop consumer groups
	for groupID, group := range cm.groups {
		if err := cm.stopConsumerGroup(ctx, group); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop consumer group %s: %w", groupID, err))
		}
	}

	// Reset stats
	cm.stats.mu.Lock()
	cm.stats.ActiveConsumers = 0
	cm.stats.ActiveGroups = 0
	cm.stats.mu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping consumers: %v", errs)
	}

	return nil
}

// startConsumerGroup starts a consumer group
func (cm *ConsumerManager) startConsumerGroup(ctx context.Context, group *ConsumerGroup) error {
	// This is a simplified implementation
	// In a real implementation, you would create multiple consumers for the group
	// and distribute topics/partitions among them

	log.Printf("Starting consumer group: %s", group.GroupID)
	return nil
}

// stopConsumerGroup stops a consumer group
func (cm *ConsumerManager) stopConsumerGroup(ctx context.Context, group *ConsumerGroup) error {
	log.Printf("Stopping consumer group: %s", group.GroupID)
	return nil
}

// GetStats returns consumer manager statistics
func (cm *ConsumerManager) GetStats() *ManagerStats {
	cm.stats.mu.RLock()
	defer cm.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &ManagerStats{
		TotalConsumers:  cm.stats.TotalConsumers,
		TotalGroups:     cm.stats.TotalGroups,
		ActiveConsumers: cm.stats.ActiveConsumers,
		ActiveGroups:    cm.stats.ActiveGroups,
		TotalMessages:   cm.stats.TotalMessages,
		FailedMessages:  cm.stats.FailedMessages,
	}

	return stats
}

// Health checks the health of all consumers
func (cm *ConsumerManager) Health(ctx context.Context) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var errs []error

	// Check individual consumers
	for consumerID, consumer := range cm.consumers {
		if err := consumer.Health(ctx); err != nil {
			errs = append(errs, fmt.Errorf("consumer %s health check failed: %w", consumerID, err))
		}
	}

	// Check consumer groups
	for groupID, group := range cm.groups {
		// Check each consumer in the group
		for i, consumer := range group.Consumers {
			if err := consumer.Health(ctx); err != nil {
				errs = append(errs, fmt.Errorf("consumer %d in group %s health check failed: %w", i, groupID, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("health check errors: %v", errs)
	}

	return nil
}

// ConsumerBuilder helps build consumers
type ConsumerBuilder struct {
	consumerID string
	topics     []string
	handler    MessageHandler
	config     map[string]interface{}
}

// NewConsumerBuilder creates a new consumer builder
func NewConsumerBuilder() *ConsumerBuilder {
	return &ConsumerBuilder{
		config: make(map[string]interface{}),
	}
}

// WithID sets the consumer ID
func (cb *ConsumerBuilder) WithID(consumerID string) *ConsumerBuilder {
	cb.consumerID = consumerID
	return cb
}

// WithTopics sets the topics to consume
func (cb *ConsumerBuilder) WithTopics(topics []string) *ConsumerBuilder {
	cb.topics = topics
	return cb
}

// WithHandler sets the message handler
func (cb *ConsumerBuilder) WithHandler(handler MessageHandler) *ConsumerBuilder {
	cb.handler = handler
	return cb
}

// WithConfig sets a configuration parameter
func (cb *ConsumerBuilder) WithConfig(key string, value interface{}) *ConsumerBuilder {
	cb.config[key] = value
	return cb
}

// Build builds the consumer configuration
func (cb *ConsumerBuilder) Build() map[string]interface{} {
	config := map[string]interface{}{
		"consumer_id": cb.consumerID,
		"topics":      cb.topics,
	}

	// Add custom config
	for k, v := range cb.config {
		config[k] = v
	}

	return config
}
