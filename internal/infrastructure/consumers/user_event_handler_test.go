package consumers_test

import (
	"context"
	"testing"

	"go-clean-ddd-es-template/internal/domain/repositories/mocks"
	"go-clean-ddd-es-template/internal/infrastructure/consumers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserEventHandler_HandleEvent(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		eventData     map[string]interface{}
		setupMocks    func(*mocks.MockUserReadRepository)
		expectedError bool
	}{
		{
			name:      "successful user.created event handling",
			eventType: "user.created",
			eventData: map[string]interface{}{
				"user_id": "user-123",
				"email":   "test@example.com",
				"name":    "John Doe",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().SaveUser(mock.Anything, mock.AnythingOfType("*entities.UserReadModel")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "successful user.updated event handling",
			eventType: "user.updated",
			eventData: map[string]interface{}{
				"user_id": "user-123",
				"email":   "updated@example.com",
				"name":    "John Updated",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().UpdateUser(mock.Anything, mock.AnythingOfType("*entities.UserReadModel")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "successful user.deleted event handling",
			eventType: "user.deleted",
			eventData: map[string]interface{}{
				"user_id": "user-123",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().DeleteUser(mock.Anything, "user-123").Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "unknown event type",
			eventType: "user.unknown",
			eventData: map[string]interface{}{
				"user_id": "user-123",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				// No expectations since event type is not handled
			},
			expectedError: false, // Should not error, just skip
		},
		{
			name:      "user creation fails",
			eventType: "user.created",
			eventData: map[string]interface{}{
				"user_id": "user-123",
				"email":   "test@example.com",
				"name":    "John Doe",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().SaveUser(mock.Anything, mock.AnythingOfType("*entities.UserReadModel")).Return(assert.AnError)
			},
			expectedError: true,
		},
		{
			name:      "user update fails",
			eventType: "user.updated",
			eventData: map[string]interface{}{
				"user_id": "user-123",
				"email":   "updated@example.com",
				"name":    "John Updated",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().UpdateUser(mock.Anything, mock.AnythingOfType("*entities.UserReadModel")).Return(assert.AnError)
			},
			expectedError: true,
		},
		{
			name:      "user deletion fails",
			eventType: "user.deleted",
			eventData: map[string]interface{}{
				"user_id": "user-123",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().DeleteUser(mock.Anything, "user-123").Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userRepo := mocks.NewMockUserReadRepository(t)

			// Setup mocks
			tt.setupMocks(userRepo)

			// Create handler
			handler := consumers.NewUserEventHandler(userRepo)

			// Handle event
			err := handler.HandleEvent(context.Background(), tt.eventType, tt.eventData)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserEventHandler_HandleEventWithInvalidData(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		eventData map[string]interface{}
	}{
		{
			name:      "missing user_id in user.created",
			eventType: "user.created",
			eventData: map[string]interface{}{
				"email": "test@example.com",
				"name":  "John Doe",
			},
		},
		{
			name:      "missing user_id in user.updated",
			eventType: "user.updated",
			eventData: map[string]interface{}{
				"email": "updated@example.com",
				"name":  "John Updated",
			},
		},
		{
			name:      "missing user_id in user.deleted",
			eventType: "user.deleted",
			eventData: map[string]interface{}{
				"other_field": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userRepo := mocks.NewMockUserReadRepository(t)

			// Create handler
			handler := consumers.NewUserEventHandler(userRepo)

			// Handle event
			err := handler.HandleEvent(context.Background(), tt.eventType, tt.eventData)

			// Should handle gracefully without error
			assert.NoError(t, err)
		})
	}
}
