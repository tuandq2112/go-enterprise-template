package services_test

import (
	"context"
	"testing"

	"go-clean-ddd-es-template/internal/application/commands"
	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/application/queries"
	"go-clean-ddd-es-template/internal/application/services"
	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/repositories/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
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
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
				eventStore.EXPECT().SaveEvent(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*events.Event")).Return(nil)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userWriteRepo := mocks.NewMockUserWriteRepository(t)
			userReadRepo := mocks.NewMockUserReadRepository(t)
			eventStore := mocks.NewMockEventStore(t)
			eventPublisher := mocks.NewMockEventPublisher(t)

			// Setup mocks
			tt.setupMocks(userWriteRepo, eventStore, eventPublisher)

			// Create command and query handlers
			createHandler := commands.NewUserCreateCommandHandler(userWriteRepo, eventStore, eventPublisher)
			updateHandler := commands.NewUserUpdateCommandHandler(userWriteRepo, eventStore, eventPublisher)
			deleteHandler := commands.NewUserDeleteCommandHandler(userWriteRepo, eventStore, eventPublisher)
			getHandler := queries.NewUserGetQueryHandler(userReadRepo)
			listHandler := queries.NewUserListQueryHandler(userReadRepo)
			getByEmailHandler := queries.NewUserGetByEmailQueryHandler(userReadRepo)
			eventsHandler := queries.NewUserEventsQueryHandler(userReadRepo)

			// Create service
			service := services.NewUserService(
				createHandler,
				updateHandler,
				deleteHandler,
				getHandler,
				listHandler,
				getByEmailHandler,
				eventsHandler,
			)

			// Execute service method
			result, err := service.CreateUser(context.Background(), tt.command)

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

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name          string
		query         dto.GetUserQuery
		setupMocks    func(*mocks.MockUserReadRepository)
		expectedError bool
	}{
		{
			name: "successful user retrieval",
			query: dto.GetUserQuery{
				UserID: "user-123",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().GetUserByID(mock.Anything, "user-123").Return(&entities.UserReadModel{
					UserID: "user-123",
					Email:  "test@example.com",
					Name:   "John Doe",
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "user not found",
			query: dto.GetUserQuery{
				UserID: "user-123",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				userRepo.EXPECT().GetUserByID(mock.Anything, "user-123").Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userWriteRepo := mocks.NewMockUserWriteRepository(t)
			userReadRepo := mocks.NewMockUserReadRepository(t)
			eventStore := mocks.NewMockEventStore(t)
			eventPublisher := mocks.NewMockEventPublisher(t)

			// Setup mocks
			tt.setupMocks(userReadRepo)

			// Create command and query handlers
			createHandler := commands.NewUserCreateCommandHandler(userWriteRepo, eventStore, eventPublisher)
			updateHandler := commands.NewUserUpdateCommandHandler(userWriteRepo, eventStore, eventPublisher)
			deleteHandler := commands.NewUserDeleteCommandHandler(userWriteRepo, eventStore, eventPublisher)
			getHandler := queries.NewUserGetQueryHandler(userReadRepo)
			listHandler := queries.NewUserListQueryHandler(userReadRepo)
			getByEmailHandler := queries.NewUserGetByEmailQueryHandler(userReadRepo)
			eventsHandler := queries.NewUserEventsQueryHandler(userReadRepo)

			// Create service
			service := services.NewUserService(
				createHandler,
				updateHandler,
				deleteHandler,
				getHandler,
				listHandler,
				getByEmailHandler,
				eventsHandler,
			)

			// Execute service method
			result, err := service.GetUser(context.Background(), tt.query)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.query.UserID, result.UserID)
			}
		})
	}
}
