package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserReadModel_Fields(t *testing.T) {
	now := time.Now()
	userID := "123e4567-e89b-12d3-a456-426614174000"

	userReadModel := UserReadModel{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Email:     "test@example.com",
		Name:      "John Doe",
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		Version:   1,
	}

	assert.NotEqual(t, primitive.NilObjectID, userReadModel.ID)
	assert.Equal(t, userID, userReadModel.UserID)
	assert.Equal(t, "test@example.com", userReadModel.Email)
	assert.Equal(t, "John Doe", userReadModel.Name)
	assert.Equal(t, now, userReadModel.CreatedAt)
	assert.Equal(t, now, userReadModel.UpdatedAt)
	assert.Nil(t, userReadModel.DeletedAt)
	assert.Equal(t, 1, userReadModel.Version)
}

func TestUserReadModel_WithDeletedAt(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(24 * time.Hour)

	userReadModel := UserReadModel{
		ID:        primitive.NewObjectID(),
		UserID:    "123e4567-e89b-12d3-a456-426614174000",
		Email:     "test@example.com",
		Name:      "John Doe",
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &deletedAt,
		Version:   1,
	}

	assert.NotNil(t, userReadModel.DeletedAt)
	assert.Equal(t, deletedAt, *userReadModel.DeletedAt)
}

func TestUserReadModel_DeletedAtField(t *testing.T) {
	now := time.Now()

	// Test non-deleted user
	userReadModel := UserReadModel{
		ID:        primitive.NewObjectID(),
		UserID:    "123e4567-e89b-12d3-a456-426614174000",
		Email:     "test@example.com",
		Name:      "John Doe",
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		Version:   1,
	}

	assert.Nil(t, userReadModel.DeletedAt)

	// Test deleted user
	deletedAt := now.Add(-1 * time.Hour) // Deleted in the past
	userReadModel.DeletedAt = &deletedAt
	assert.NotNil(t, userReadModel.DeletedAt)
	assert.Equal(t, deletedAt, *userReadModel.DeletedAt)
}

func TestUserEvent_Fields(t *testing.T) {
	now := time.Now()
	userID := "123e4567-e89b-12d3-a456-426614174000"
	eventData := map[string]interface{}{
		"email": "test@example.com",
		"name":  "John Doe",
	}

	userEvent := UserEvent{
		UserID:    userID,
		EventType: "user.created",
		EventData: eventData,
		Timestamp: now,
		Version:   1,
	}

	assert.Equal(t, userID, userEvent.UserID)
	assert.Equal(t, "user.created", userEvent.EventType)
	assert.Equal(t, eventData, userEvent.EventData)
	assert.Equal(t, now, userEvent.Timestamp)
	assert.Equal(t, 1, userEvent.Version)
}

func TestUserEvent_EventDataAccess(t *testing.T) {
	eventData := map[string]interface{}{
		"email": "test@example.com",
		"name":  "John Doe",
		"age":   30,
	}

	userEvent := UserEvent{
		UserID:    "123e4567-e89b-12d3-a456-426614174000",
		EventType: "user.created",
		EventData: eventData,
		Timestamp: time.Now(),
		Version:   1,
	}

	// Test accessing event data
	assert.Equal(t, "test@example.com", userEvent.EventData["email"])
	assert.Equal(t, "John Doe", userEvent.EventData["name"])
	assert.Equal(t, 30, userEvent.EventData["age"])
}

func TestUserSummary_Fields(t *testing.T) {
	now := time.Now()
	userID := "123e4567-e89b-12d3-a456-426614174000"

	userSummary := UserSummary{
		UserID:    userID,
		Email:     "test@example.com",
		Name:      "John Doe",
		CreatedAt: now,
	}

	assert.Equal(t, userID, userSummary.UserID)
	assert.Equal(t, "test@example.com", userSummary.Email)
	assert.Equal(t, "John Doe", userSummary.Name)
	assert.Equal(t, now, userSummary.CreatedAt)
}

func TestUserReadModel_FromUser(t *testing.T) {
	// Create a user entity
	user, err := NewUser("test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Create UserReadModel from user
	now := time.Now()
	userReadModel := UserReadModel{
		ID:        primitive.NewObjectID(),
		UserID:    user.GetID(),
		Email:     user.GetEmail(),
		Name:      user.GetName(),
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	assert.Equal(t, user.GetID(), userReadModel.UserID)
	assert.Equal(t, user.GetEmail(), userReadModel.Email)
	assert.Equal(t, user.GetName(), userReadModel.Name)
	assert.Equal(t, 1, userReadModel.Version)
}

func TestUserEvent_FromEvent(t *testing.T) {
	// Create event data
	eventData := map[string]interface{}{
		"user_id": "123e4567-e89b-12d3-a456-426614174000",
		"email":   "test@example.com",
		"name":    "John Doe",
	}

	// Create UserEvent
	now := time.Now()
	userEvent := UserEvent{
		UserID:    "123e4567-e89b-12d3-a456-426614174000",
		EventType: "user.created",
		EventData: eventData,
		Timestamp: now,
		Version:   1,
	}

	assert.Equal(t, "user.created", userEvent.EventType)
	assert.Equal(t, eventData, userEvent.EventData)
	assert.Equal(t, now, userEvent.Timestamp)
	assert.Equal(t, 1, userEvent.Version)
}

func TestUserSummary_FromUserReadModel(t *testing.T) {
	// Create UserReadModel
	now := time.Now()
	userReadModel := UserReadModel{
		ID:        primitive.NewObjectID(),
		UserID:    "123e4567-e89b-12d3-a456-426614174000",
		Email:     "test@example.com",
		Name:      "John Doe",
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	// Create UserSummary from UserReadModel
	userSummary := UserSummary{
		UserID:    userReadModel.UserID,
		Email:     userReadModel.Email,
		Name:      userReadModel.Name,
		CreatedAt: userReadModel.CreatedAt,
	}

	assert.Equal(t, userReadModel.UserID, userSummary.UserID)
	assert.Equal(t, userReadModel.Email, userSummary.Email)
	assert.Equal(t, userReadModel.Name, userSummary.Name)
	assert.Equal(t, userReadModel.CreatedAt, userSummary.CreatedAt)
}
