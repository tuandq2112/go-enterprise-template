package queries

import (
	"context"
	"testing"
	"time"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/repositories/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserGetQueryHandler_Handle(t *testing.T) {
	tests := []struct {
		name          string
		query         dto.GetUserQuery
		setupMocks    func(*mocks.MockUserReadRepository)
		expectedError bool
		expectedUser  *dto.GetUserQueryResponse
	}{
		{
			name: "successful user retrieval",
			query: dto.GetUserQuery{
				UserID: "user-123",
			},
			setupMocks: func(userRepo *mocks.MockUserReadRepository) {
				now := time.Now()
				userRepo.EXPECT().GetUserByID(mock.Anything, "user-123").Return(&entities.UserReadModel{
					ID:        primitive.NewObjectID(),
					UserID:    "user-123",
					Email:     "test@example.com",
					Name:      "John Doe",
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				}, nil)
			},
			expectedError: false,
			expectedUser: &dto.GetUserQueryResponse{
				UserID:    "user-123",
				Email:     "test@example.com",
				Name:      "John Doe",
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
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
			expectedUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			userRepo := mocks.NewMockUserReadRepository(t)

			// Setup mocks
			tt.setupMocks(userRepo)

			// Create handler
			handler := NewUserGetQueryHandler(userRepo)

			// Execute query
			result, err := handler.Handle(context.Background(), tt.query)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.UserID, result.UserID)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
				assert.Equal(t, tt.expectedUser.Name, result.Name)
			}
		})
	}
}
