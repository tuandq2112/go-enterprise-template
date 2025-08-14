package commands

import (
	"context"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/domain/repositories"
)

// UserUpdateCommandHandler handles the update user command (write operation)
type UserUpdateCommandHandler struct {
	userWriteRepo  repositories.UserWriteRepository
	eventStore     repositories.EventStore
	eventPublisher repositories.EventPublisher
}

// NewUserUpdateCommandHandler creates a new user update command handler
func NewUserUpdateCommandHandler(
	userWriteRepo repositories.UserWriteRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
) *UserUpdateCommandHandler {
	return &UserUpdateCommandHandler{
		userWriteRepo:  userWriteRepo,
		eventStore:     eventStore,
		eventPublisher: eventPublisher,
	}
}

// Handle handles the update user command
func (h *UserUpdateCommandHandler) Handle(ctx context.Context, cmd dto.UpdateUserCommand) (*dto.UpdateUserCommandResponse, error) {
	// Get existing user from write database
	user, err := h.userWriteRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	// Update user with validation
	if err := user.UpdateName(cmd.Name); err != nil {
		return nil, err
	}

	// Save to write database (PostgreSQL)
	if err := h.userWriteRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Create domain event
	userUpdatedEvent := &events.UserUpdatedEvent{
		UserID:    user.GetID(),
		Name:      user.GetName(),
		UpdatedAt: user.UpdatedAt,
	}

	// Wrap in Event
	event, err := events.NewEvent("user.updated", userUpdatedEvent, 1)
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
	response := &dto.UpdateUserCommandResponse{
		UserID:    user.GetID(),
		Name:      user.GetName(),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return response, nil
}
