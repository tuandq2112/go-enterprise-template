package consumer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// KafkaConsumer implements Consumer interface for Kafka
type KafkaConsumer struct {
	consumer sarama.Consumer
	groupID  string
	topics   []string
	handlers map[string]MessageHandler
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
	stats    *ConsumerStats
	config   *KafkaConsumerConfig
}

// KafkaConsumerConfig holds Kafka consumer configuration
type KafkaConsumerConfig struct {
	Brokers            []string
	GroupID            string
	Topics             []string
	AutoCommit         bool
	AutoCommitInterval time.Duration
	SessionTimeout     time.Duration
	HeartbeatInterval  time.Duration
	MaxPollRecords     int
	MaxPollInterval    time.Duration
	OffsetReset        string // "earliest", "latest"
	WorkerPoolSize     int
}

// DefaultKafkaConsumerConfig returns default Kafka consumer configuration
func DefaultKafkaConsumerConfig() *KafkaConsumerConfig {
	return &KafkaConsumerConfig{
		Brokers:            []string{"localhost:9092"},
		GroupID:            "default-group",
		AutoCommit:         true,
		AutoCommitInterval: 5 * time.Second,
		SessionTimeout:     30 * time.Second,
		HeartbeatInterval:  3 * time.Second,
		MaxPollRecords:     500,
		MaxPollInterval:    300 * time.Second,
		OffsetReset:        "latest",
		WorkerPoolSize:     10,
	}
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(config *KafkaConsumerConfig) (*KafkaConsumer, error) {
	if config == nil {
		config = DefaultKafkaConsumerConfig()
	}

	// Create Sarama consumer config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	if config.OffsetReset == "earliest" {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = config.AutoCommit
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = config.AutoCommitInterval
	saramaConfig.Consumer.Group.Session.Timeout = config.SessionTimeout
	saramaConfig.Consumer.Group.Heartbeat.Interval = config.HeartbeatInterval
	saramaConfig.Consumer.MaxWaitTime = config.MaxPollInterval
	saramaConfig.Consumer.Fetch.Max = int32(config.MaxPollRecords)

	// Create Sarama consumer
	consumer, err := sarama.NewConsumer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	kafkaConsumer := &KafkaConsumer{
		consumer: consumer,
		groupID:  config.GroupID,
		topics:   config.Topics,
		handlers: make(map[string]MessageHandler),
		stopChan: make(chan struct{}),
		stats:    &ConsumerStats{ConsumerLag: make(map[string]int64)},
		config:   config,
	}

	return kafkaConsumer, nil
}

// Start starts the Kafka consumer
func (kc *KafkaConsumer) Start(ctx context.Context) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if kc.running {
		return fmt.Errorf("consumer is already running")
	}

	kc.running = true
	kc.stats.mu.Lock()
	kc.stats.IsRunning = true
	kc.stats.mu.Unlock()

	log.Printf("Starting Kafka consumer for group: %s", kc.groupID)

	// Start consuming from each topic
	for _, topic := range kc.topics {
		kc.wg.Add(1)
		go kc.consumeTopic(ctx, topic)
	}

	return nil
}

// Stop stops the Kafka consumer
func (kc *KafkaConsumer) Stop(ctx context.Context) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if !kc.running {
		return nil
	}

	log.Printf("Stopping Kafka consumer for group: %s", kc.groupID)

	kc.running = false
	close(kc.stopChan)
	kc.wg.Wait()

	// Close Sarama consumer
	if err := kc.consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	kc.stats.mu.Lock()
	kc.stats.IsRunning = false
	kc.stats.mu.Unlock()

	return nil
}

// IsRunning returns whether the consumer is running
func (kc *KafkaConsumer) IsRunning() bool {
	kc.mu.RLock()
	defer kc.mu.RUnlock()
	return kc.running
}

// Subscribe subscribes to a topic with a handler
func (kc *KafkaConsumer) Subscribe(topic string, handler MessageHandler) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	kc.handlers[topic] = handler
	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

// Unsubscribe unsubscribes from a topic
func (kc *KafkaConsumer) Unsubscribe(topic string) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	delete(kc.handlers, topic)
	log.Printf("Unsubscribed from topic: %s", topic)
	return nil
}

// JoinGroup joins a consumer group (not implemented for simple consumer)
func (kc *KafkaConsumer) JoinGroup(groupID string) error {
	// This is a simplified implementation
	// For consumer groups, you would use sarama.ConsumerGroup
	log.Printf("Joined consumer group: %s", groupID)
	return nil
}

// LeaveGroup leaves the consumer group
func (kc *KafkaConsumer) LeaveGroup() error {
	// This is a simplified implementation
	log.Printf("Left consumer group: %s", kc.groupID)
	return nil
}

// Health checks the health of the consumer
func (kc *KafkaConsumer) Health(ctx context.Context) error {
	if !kc.IsRunning() {
		return fmt.Errorf("consumer is not running")
	}

	// Check if consumer is still valid
	if kc.consumer == nil {
		return fmt.Errorf("consumer is nil")
	}

	return nil
}

// GetStats returns consumer statistics
func (kc *KafkaConsumer) GetStats(ctx context.Context) (*ConsumerStats, error) {
	kc.stats.mu.RLock()
	defer kc.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &ConsumerStats{
		GroupID:          kc.stats.GroupID,
		Topics:           kc.stats.Topics,
		MessagesConsumed: kc.stats.MessagesConsumed,
		MessagesFailed:   kc.stats.MessagesFailed,
		MessagesRetried:  kc.stats.MessagesRetried,
		LastMessageTime:  kc.stats.LastMessageTime,
		ConsumerLag:      make(map[string]int64),
		ActiveConsumers:  kc.stats.ActiveConsumers,
		IsRunning:        kc.stats.IsRunning,
	}

	// Copy consumer lag
	for topic, lag := range kc.stats.ConsumerLag {
		stats.ConsumerLag[topic] = lag
	}

	return stats, nil
}

// consumeTopic consumes messages from a specific topic
func (kc *KafkaConsumer) consumeTopic(ctx context.Context, topic string) {
	defer kc.wg.Done()

	// Get partition list for the topic
	partitions, err := kc.consumer.Partitions(topic)
	if err != nil {
		log.Printf("[ERROR] Failed to get partitions for topic %s: %v", topic, err)
		return
	}

	// Create partition consumers
	for _, partition := range partitions {
		partitionConsumer, err := kc.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
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
			case <-kc.stopChan:
				log.Printf("[INFO] Stop signal received, stopping consumer for topic %s partition %d", topic, partition)
				return
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					kc.handleMessage(ctx, topic, partition, msg)
				}
			case err := <-partitionConsumer.Errors():
				if err != nil {
					log.Printf("[ERROR] Error consuming from topic %s partition %d: %v", topic, partition, err)
					kc.incrementFailedMessages()
				}
			}
		}
	}
}

// handleMessage handles a single message
func (kc *KafkaConsumer) handleMessage(ctx context.Context, topic string, partition int32, msg *sarama.ConsumerMessage) {
	// Convert Sarama message to our Message type
	message := &Message{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   make(map[string][]byte),
		Timestamp: msg.Timestamp,
	}

	// Convert headers
	for _, header := range msg.Headers {
		message.Headers[string(header.Key)] = header.Value
	}

	// Get handler for topic
	kc.mu.RLock()
	handler, exists := kc.handlers[topic]
	kc.mu.RUnlock()

	if !exists {
		log.Printf("[WARN] No handler registered for topic: %s", topic)
		return
	}

	// Process message with retry logic
	err := kc.processMessageWithRetry(ctx, handler, message)
	if err != nil {
		log.Printf("[ERROR] Failed to process message from topic %s partition %d offset %d: %v",
			topic, partition, msg.Offset, err)
		kc.incrementFailedMessages()
	} else {
		kc.incrementConsumedMessages()
		log.Printf("[INFO] Successfully processed message from topic %s partition %d offset %d",
			topic, partition, msg.Offset)
	}
}

// processMessageWithRetry processes a message with retry logic
func (kc *KafkaConsumer) processMessageWithRetry(ctx context.Context, handler MessageHandler, message *Message) error {
	maxRetries := 3
	delay := time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := handler(ctx, message); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxRetries {
				kc.incrementRetriedMessages()
				log.Printf("[WARN] Attempt %d failed, retrying in %v: %v", attempt, delay, err)
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// incrementConsumedMessages increments the consumed messages count
func (kc *KafkaConsumer) incrementConsumedMessages() {
	kc.stats.mu.Lock()
	defer kc.stats.mu.Unlock()
	kc.stats.MessagesConsumed++
	kc.stats.LastMessageTime = time.Now()
}

// incrementFailedMessages increments the failed messages count
func (kc *KafkaConsumer) incrementFailedMessages() {
	kc.stats.mu.Lock()
	defer kc.stats.mu.Unlock()
	kc.stats.MessagesFailed++
}

// incrementRetriedMessages increments the retried messages count
func (kc *KafkaConsumer) incrementRetriedMessages() {
	kc.stats.mu.Lock()
	defer kc.stats.mu.Unlock()
	kc.stats.MessagesRetried++
}

// KafkaConsumerGroup implements consumer group functionality
type KafkaConsumerGroup struct {
	group    sarama.ConsumerGroup
	handlers map[string]MessageHandler
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
	stats    *ConsumerStats
	config   *KafkaConsumerConfig
}

// NewKafkaConsumerGroup creates a new Kafka consumer group
func NewKafkaConsumerGroup(config *KafkaConsumerConfig) (*KafkaConsumerGroup, error) {
	if config == nil {
		config = DefaultKafkaConsumerConfig()
	}

	// Create Sarama consumer group config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	if config.OffsetReset == "earliest" {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = config.AutoCommit
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = config.AutoCommitInterval
	saramaConfig.Consumer.Group.Session.Timeout = config.SessionTimeout
	saramaConfig.Consumer.Group.Heartbeat.Interval = config.HeartbeatInterval

	// Create Sarama consumer group
	group, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	kafkaGroup := &KafkaConsumerGroup{
		group:    group,
		handlers: make(map[string]MessageHandler),
		stopChan: make(chan struct{}),
		stats:    &ConsumerStats{ConsumerLag: make(map[string]int64)},
		config:   config,
	}

	return kafkaGroup, nil
}

// Start starts the Kafka consumer group
func (kcg *KafkaConsumerGroup) Start(ctx context.Context) error {
	kcg.mu.Lock()
	defer kcg.mu.Unlock()

	if kcg.running {
		return fmt.Errorf("consumer group is already running")
	}

	kcg.running = true
	kcg.stats.mu.Lock()
	kcg.stats.IsRunning = true
	kcg.stats.mu.Unlock()

	log.Printf("Starting Kafka consumer group: %s", kcg.config.GroupID)

	// Start consuming
	kcg.wg.Add(1)
	go func() {
		defer kcg.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-kcg.stopChan:
				return
			default:
				if err := kcg.group.Consume(ctx, kcg.config.Topics, kcg); err != nil {
					log.Printf("[ERROR] Error from consumer group: %v", err)
				}
			}
		}
	}()

	return nil
}

// Stop stops the Kafka consumer group
func (kcg *KafkaConsumerGroup) Stop(ctx context.Context) error {
	kcg.mu.Lock()
	defer kcg.mu.Unlock()

	if !kcg.running {
		return nil
	}

	log.Printf("Stopping Kafka consumer group: %s", kcg.config.GroupID)

	kcg.running = false
	close(kcg.stopChan)
	kcg.wg.Wait()

	// Close Sarama consumer group
	if err := kcg.group.Close(); err != nil {
		log.Printf("Error closing Kafka consumer group: %v", err)
	}

	kcg.stats.mu.Lock()
	kcg.stats.IsRunning = false
	kcg.stats.mu.Unlock()

	return nil
}

// IsRunning returns whether the consumer group is running
func (kcg *KafkaConsumerGroup) IsRunning() bool {
	kcg.mu.RLock()
	defer kcg.mu.RUnlock()
	return kcg.running
}

// Subscribe subscribes to a topic with a handler
func (kcg *KafkaConsumerGroup) Subscribe(topic string, handler MessageHandler) error {
	kcg.mu.Lock()
	defer kcg.mu.Unlock()

	kcg.handlers[topic] = handler
	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

// Unsubscribe unsubscribes from a topic
func (kcg *KafkaConsumerGroup) Unsubscribe(topic string) error {
	kcg.mu.Lock()
	defer kcg.mu.Unlock()

	delete(kcg.handlers, topic)
	log.Printf("Unsubscribed from topic: %s", topic)
	return nil
}

// JoinGroup joins a consumer group
func (kcg *KafkaConsumerGroup) JoinGroup(groupID string) error {
	// Already joined when created
	log.Printf("Joined consumer group: %s", groupID)
	return nil
}

// LeaveGroup leaves the consumer group
func (kcg *KafkaConsumerGroup) LeaveGroup() error {
	log.Printf("Left consumer group: %s", kcg.config.GroupID)
	return nil
}

// Health checks the health of the consumer group
func (kcg *KafkaConsumerGroup) Health(ctx context.Context) error {
	if !kcg.IsRunning() {
		return fmt.Errorf("consumer group is not running")
	}

	if kcg.group == nil {
		return fmt.Errorf("consumer group is nil")
	}

	return nil
}

// GetStats returns consumer group statistics
func (kcg *KafkaConsumerGroup) GetStats(ctx context.Context) (*ConsumerStats, error) {
	kcg.stats.mu.RLock()
	defer kcg.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &ConsumerStats{
		GroupID:          kcg.stats.GroupID,
		Topics:           kcg.stats.Topics,
		MessagesConsumed: kcg.stats.MessagesConsumed,
		MessagesFailed:   kcg.stats.MessagesFailed,
		MessagesRetried:  kcg.stats.MessagesRetried,
		LastMessageTime:  kcg.stats.LastMessageTime,
		ConsumerLag:      make(map[string]int64),
		ActiveConsumers:  kcg.stats.ActiveConsumers,
		IsRunning:        kcg.stats.IsRunning,
	}

	// Copy consumer lag
	for topic, lag := range kcg.stats.ConsumerLag {
		stats.ConsumerLag[topic] = lag
	}

	return stats, nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler
func (kcg *KafkaConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-kcg.stopChan:
			return nil
		case msg := <-claim.Messages():
			if msg != nil {
				kcg.handleMessage(context.Background(), msg.Topic, msg.Partition, msg)
			}
		}
	}
}

// Setup implements sarama.ConsumerGroupHandler
func (kcg *KafkaConsumerGroup) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (kcg *KafkaConsumerGroup) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// handleMessage handles a single message (same as KafkaConsumer)
func (kcg *KafkaConsumerGroup) handleMessage(ctx context.Context, topic string, partition int32, msg *sarama.ConsumerMessage) {
	// Convert Sarama message to our Message type
	message := &Message{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   make(map[string][]byte),
		Timestamp: msg.Timestamp,
	}

	// Convert headers
	for _, header := range msg.Headers {
		message.Headers[string(header.Key)] = header.Value
	}

	// Get handler for topic
	kcg.mu.RLock()
	handler, exists := kcg.handlers[topic]
	kcg.mu.RUnlock()

	if !exists {
		log.Printf("[WARN] No handler registered for topic: %s", topic)
		return
	}

	// Process message with retry logic
	err := kcg.processMessageWithRetry(ctx, handler, message)
	if err != nil {
		log.Printf("[ERROR] Failed to process message from topic %s partition %d offset %d: %v",
			topic, partition, msg.Offset, err)
		kcg.incrementFailedMessages()
	} else {
		kcg.incrementConsumedMessages()
		log.Printf("[INFO] Successfully processed message from topic %s partition %d offset %d",
			topic, partition, msg.Offset)
	}
}

// processMessageWithRetry processes a message with retry logic (same as KafkaConsumer)
func (kcg *KafkaConsumerGroup) processMessageWithRetry(ctx context.Context, handler MessageHandler, message *Message) error {
	maxRetries := 3
	delay := time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := handler(ctx, message); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxRetries {
				kcg.incrementRetriedMessages()
				log.Printf("[WARN] Attempt %d failed, retrying in %v: %v", attempt, delay, err)
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// incrementConsumedMessages increments the consumed messages count
func (kcg *KafkaConsumerGroup) incrementConsumedMessages() {
	kcg.stats.mu.Lock()
	defer kcg.stats.mu.Unlock()
	kcg.stats.MessagesConsumed++
	kcg.stats.LastMessageTime = time.Now()
}

// incrementFailedMessages increments the failed messages count
func (kcg *KafkaConsumerGroup) incrementFailedMessages() {
	kcg.stats.mu.Lock()
	defer kcg.stats.mu.Unlock()
	kcg.stats.MessagesFailed++
}

// incrementRetriedMessages increments the retried messages count
func (kcg *KafkaConsumerGroup) incrementRetriedMessages() {
	kcg.stats.mu.Lock()
	defer kcg.stats.mu.Unlock()
	kcg.stats.MessagesRetried++
}
