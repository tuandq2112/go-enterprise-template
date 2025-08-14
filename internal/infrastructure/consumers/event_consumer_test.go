package consumers

import (
	"context"
	"testing"

	"go-clean-ddd-es-template/internal/infrastructure/consumers/mocks"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSaramaConsumer implements sarama.Consumer for testing
type MockSaramaConsumer struct {
	mock.Mock
}

func (m *MockSaramaConsumer) Partitions(topic string) ([]int32, error) {
	args := m.Called(topic)
	return args.Get(0).([]int32), args.Error(1)
}

func (m *MockSaramaConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	args := m.Called(topic, partition, offset)
	return args.Get(0).(sarama.PartitionConsumer), args.Error(1)
}

func (m *MockSaramaConsumer) Topics() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockSaramaConsumer) HighWaterMarks() map[string]map[int32]int64 {
	args := m.Called()
	return args.Get(0).(map[string]map[int32]int64)
}

func (m *MockSaramaConsumer) Pause(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockSaramaConsumer) Resume(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockSaramaConsumer) PauseAll() {
	m.Called()
}

func (m *MockSaramaConsumer) ResumeAll() {
	m.Called()
}

func (m *MockSaramaConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestEventConsumer_RegisterEventHandler(t *testing.T) {
	// Create mocks
	consumer := &MockSaramaConsumer{}
	eventHandler := mocks.NewMockEventHandler(t)

	// Create event consumer
	eventConsumer := NewEventConsumer(consumer, "test-group", []string{"test-topic"})

	// Test registering event handler
	eventConsumer.RegisterEventHandler("user.created", eventHandler)

	// Verify handler is registered by testing handleMessage
	ctx := context.Background()
	msg := &sarama.ConsumerMessage{
		Topic: "test-topic",
		Value: []byte(`{"type": "user.created", "data": {"user_id": "123"}}`),
	}

	// Setup mock expectation
	eventHandler.EXPECT().HandleEvent(ctx, "user.created", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Test handling message
	err := eventConsumer.handleMessage(ctx, msg)
	assert.NoError(t, err)
}

func TestEventConsumer_HandleMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       *sarama.ConsumerMessage
		setupMocks    func(*mocks.MockEventHandler)
		expectedError bool
	}{
		{
			name: "successful event handling",
			message: &sarama.ConsumerMessage{
				Topic: "test-topic",
				Value: []byte(`{"type": "user.created", "data": {"user_id": "123", "email": "test@example.com"}}`),
			},
			setupMocks: func(handler *mocks.MockEventHandler) {
				handler.EXPECT().HandleEvent(mock.Anything, "user.created", mock.AnythingOfType("map[string]interface {}")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "event handler not found",
			message: &sarama.ConsumerMessage{
				Topic: "test-topic",
				Value: []byte(`{"type": "unknown.event", "data": {"data": "test"}}`),
			},
			setupMocks: func(handler *mocks.MockEventHandler) {
				// No expectations since handler is not registered
			},
			expectedError: false, // Should not error, just skip
		},
		{
			name: "event handler fails",
			message: &sarama.ConsumerMessage{
				Topic: "test-topic",
				Value: []byte(`{"type": "user.created", "data": {"user_id": "123"}}`),
			},
			setupMocks: func(handler *mocks.MockEventHandler) {
				handler.EXPECT().HandleEvent(mock.Anything, "user.created", mock.AnythingOfType("map[string]interface {}")).Return(assert.AnError)
			},
			expectedError: true,
		},
		{
			name: "invalid message format",
			message: &sarama.ConsumerMessage{
				Topic: "test-topic",
				Value: []byte(`invalid json`),
			},
			setupMocks: func(handler *mocks.MockEventHandler) {
				// No expectations since message parsing will fail
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			consumer := &MockSaramaConsumer{}
			eventHandler := mocks.NewMockEventHandler(t)

			// Create event consumer
			eventConsumer := NewEventConsumer(consumer, "test-group", []string{"test-topic"})

			// Register handler for user.created events
			eventConsumer.RegisterEventHandler("user.created", eventHandler)

			// Setup mocks
			tt.setupMocks(eventHandler)

			// Handle message
			err := eventConsumer.handleMessage(context.Background(), tt.message)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEventConsumer_GetTopics(t *testing.T) {
	// Create mocks
	consumer := &MockSaramaConsumer{}
	topics := []string{"user-events", "product-events"}

	// Create event consumer
	eventConsumer := NewEventConsumer(consumer, "test-group", topics)

	// Get topics by accessing the field directly for testing
	// In a real implementation, you might want to add a getter method
	assert.Equal(t, topics, eventConsumer.topics)
}

func TestEventConsumer_GetConsumerGroup(t *testing.T) {
	// Create mocks
	consumer := &MockSaramaConsumer{}
	groupID := "test-group"

	// Create event consumer
	eventConsumer := NewEventConsumer(consumer, groupID, []string{"test-topic"})

	// Get group ID by accessing the field directly for testing
	// In a real implementation, you might want to add a getter method
	assert.Equal(t, groupID, eventConsumer.consumerGroup)
}
