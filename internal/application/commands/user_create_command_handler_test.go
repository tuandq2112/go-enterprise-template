package commands

import (
	"context"
	"testing"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/repositories/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserCreateCommandHandler_Handle(t *testing.T) {
	tests := []struct {
		name          string
		command       dto.CreateUserCommand
		setupMocks    func(*mocks.MockUserWriteRepository, *mocks.MockEventStore, *mocks.MockEventPublisher)
		expectedError bool
	}{
		{
			name: "successful user creation",
			command: dto.CreateUserCommand{
				Email: "test@example.com",
				Name:  "John Doe",
			},
			setupMocks: func(userRepo *mocks.MockUserWriteRepository, eventStore *mocks.MockEventStore, eventPublisher *mocks.MockEventPublisher) {
				// Mock user creation
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)

				// Mock event storage
				eventStore.EXPECT().SaveEvent(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*events.Event")).Return(nil)

				// Mock event publishing
				eventPublisher.EXPECT().PublishEvent(mock.Anything, mock.AnythingOfType("*events.Event")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "user creation fails",
			command: dto.CreateUserCommand{
				Email: "test@example.com",
				Name:  "John Doe",
			},
			setupMocks: func(userRepo *mocks.MockUserWriteRepository, eventStore *mocks.MockEventStore, eventPublisher *mocks.MockEventPublisher) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.User")).Return(assert.AnError)
			},
			expectedError: true,
		},
		{
			name: "event storage fails",
			command: dto.CreateUserCommand{
				Email: "test@example.com",
				Name:  "John Doe",
			},
			setupMocks: func(userRepo *mocks.MockUserWriteRepository, eventStore *mocks.MockEventStore, eventPublisher *mocks.MockEventPublisher) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)

				eventStore.EXPECT().SaveEvent(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*events.Event")).Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userRepo := mocks.NewMockUserWriteRepository(t)
			eventStore := mocks.NewMockEventStore(t)
			eventPublisher := mocks.NewMockEventPublisher(t)

			// Setup mocks
			tt.setupMocks(userRepo, eventStore, eventPublisher)

			// Create handler
			handler := NewUserCreateCommandHandler(userRepo, eventStore, eventPublisher)

			// Execute command
			result, err := handler.Handle(context.Background(), tt.command)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.command.Email, result.Email)
				assert.Equal(t, tt.command.Name, result.Name)
			}
		})
	}
}
