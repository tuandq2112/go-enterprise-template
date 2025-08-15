package consumers

import (
	"context"
	"fmt"
	"time"

	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/repositories"
)

// UserEventHandler handles user-specific events
type UserEventHandler struct {
	readRepository repositories.UserReadRepository
}

// NewUserEventHandler creates a new user event handler
func NewUserEventHandler(readRepository repositories.UserReadRepository) *UserEventHandler {
	return &UserEventHandler{
		readRepository: readRepository,
	}
}

// HandleEvent handles user events
func (h *UserEventHandler) HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error {
	switch eventType {
	case "user.created":
		return h.handleUserCreated(ctx, eventData)
	case "user.updated":
		return h.handleUserUpdated(ctx, eventData)
	case "user.deleted":
		return h.handleUserDeleted(ctx, eventData)
	default:
		return fmt.Errorf("unknown user event type: %s", eventType)
	}
}

// handleUserCreated handles user.created event
func (h *UserEventHandler) handleUserCreated(ctx context.Context, data map[string]interface{}) error {
	userID, _ := data["user_id"].(string)
	email, _ := data["email"].(string)
	name, _ := data["name"].(string)
	createdAtStr, _ := data["created_at"].(string)

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		createdAt = time.Now()
	}

	// Create read model
	userReadModel := &entities.UserReadModel{
		UserID:    userID,
		Email:     email,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		Version:   1,
	}

	// Save to MongoDB
	if err := h.readRepository.SaveUser(ctx, userReadModel); err != nil {
		return err
	}

	// Save event to MongoDB
	userEvent := &entities.UserEvent{
		UserID:    userID,
		EventType: "user.created",
		EventData: data,
		Timestamp: time.Now(),
		Version:   1,
	}

	if err := h.readRepository.SaveEvent(ctx, userEvent); err != nil {
		return err
	}

	return nil
}

// handleUserUpdated handles user.updated event
func (h *UserEventHandler) handleUserUpdated(ctx context.Context, data map[string]interface{}) error {
	userID, _ := data["user_id"].(string)
	name, _ := data["name"].(string)
	updatedAtStr, _ := data["updated_at"].(string)

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		updatedAt = time.Now()
	}

	// Get existing user from MongoDB
	existingUser, err := h.readRepository.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Update user
	existingUser.Name = name
	existingUser.UpdatedAt = updatedAt
	existingUser.Version++

	// Save to MongoDB
	if err := h.readRepository.UpdateUser(ctx, existingUser); err != nil {
		return err
	}

	// Save event to MongoDB
	userEvent := &entities.UserEvent{
		UserID:    userID,
		EventType: "user.updated",
		EventData: data,
		Timestamp: time.Now(),
		Version:   existingUser.Version,
	}

	if err := h.readRepository.SaveEvent(ctx, userEvent); err != nil {
		return err
	}

	return nil
}

// handleUserDeleted handles user.deleted event
func (h *UserEventHandler) handleUserDeleted(ctx context.Context, data map[string]interface{}) error {
	userID, _ := data["user_id"].(string)
	deletedAtStr, _ := data["deleted_at"].(string)

	deletedAt, err := time.Parse(time.RFC3339, deletedAtStr)
	if err != nil {
		deletedAt = time.Now()
	}

	// Get existing user from MongoDB
	existingUser, err := h.readRepository.GetUserByID(ctx, userID)
	if err != nil {
		// If user doesn't exist, create a minimal user record for deletion
		existingUser = &entities.UserReadModel{
			UserID:    userID,
			Email:     "", // Will be filled from event data if available
			Name:      "", // Will be filled from event data if available
			CreatedAt: deletedAt,
			UpdatedAt: deletedAt,
			Version:   1,
		}

		// Try to get email and name from event data
		if email, ok := data["email"].(string); ok {
			existingUser.Email = email
		}
		if name, ok := data["name"].(string); ok {
			existingUser.Name = name
		}
	}

	// Soft delete user
	existingUser.DeletedAt = &deletedAt
	existingUser.UpdatedAt = deletedAt
	existingUser.Version++

	// Save to MongoDB (create if not exists, update if exists)
	if err := h.readRepository.UpdateUser(ctx, existingUser); err != nil {
		// If update fails, try to create
		if err := h.readRepository.SaveUser(ctx, existingUser); err != nil {
			return err
		}
	}

	// Save event to MongoDB
	userEvent := &entities.UserEvent{
		UserID:    userID,
		EventType: "user.deleted",
		EventData: data,
		Timestamp: time.Now(),
		Version:   existingUser.Version,
	}

	if err := h.readRepository.SaveEvent(ctx, userEvent); err != nil {
		return err
	}

	return nil
}
