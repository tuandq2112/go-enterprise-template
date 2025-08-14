package commands

import (
	"context"
	"time"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/domain/repositories"
)

// UserDeleteCommandHandler handles the delete user command (write operation)
type UserDeleteCommandHandler struct {
	userWriteRepo  repositories.UserWriteRepository
	eventStore     repositories.EventStore
	eventPublisher repositories.EventPublisher
}

// NewUserDeleteCommandHandler creates a new user delete command handler
func NewUserDeleteCommandHandler(
	userWriteRepo repositories.UserWriteRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
) *UserDeleteCommandHandler {
	return &UserDeleteCommandHandler{
		userWriteRepo:  userWriteRepo,
		eventStore:     eventStore,
		eventPublisher: eventPublisher,
	}
}

// Handle handles the delete user command
func (h *UserDeleteCommandHandler) Handle(ctx context.Context, cmd dto.DeleteUserCommand) (*dto.DeleteUserCommandResponse, error) {
	// Get existing user from write database
	user, err := h.userWriteRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	// Delete from write database (PostgreSQL)
	if err := h.userWriteRepo.Delete(ctx, cmd.UserID); err != nil {
		return nil, err
	}

	// Create domain event
	userDeletedEvent := &events.UserDeletedEvent{
		UserID:    user.GetID(),
		DeletedAt: time.Now(),
	}

	// Wrap in Event
	event, err := events.NewEvent("user.deleted", userDeletedEvent, 1)
	if err != nil {
		return nil, err
	}

	// Save event to event store
	if err := h.eventStore.SaveEvent(ctx, user.GetID(), event); err != nil {
		return nil, err
	}

	// Publish event to Kafka
	if err := h.eventPublisher.PublishEvent(ctx, event); err != nil {
		return nil, err
	}

	// Return response
	response := &dto.DeleteUserCommandResponse{
		UserID:    user.GetID(),
		DeletedAt: userDeletedEvent.DeletedAt.Format("2006-01-02T15:04:05Z07:00"),
		Success:   true,
	}

	return response, nil
}
