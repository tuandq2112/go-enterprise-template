package commands

import (
	"context"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/domain/repositories"
	"go-clean-ddd-es-template/pkg/errors"
)

// UserCreateCommandHandler handles the create user command (write operation)
type UserCreateCommandHandler struct {
	userWriteRepo  repositories.UserWriteRepository
	eventStore     repositories.EventStore
	eventPublisher repositories.EventPublisher
}

// NewUserCreateCommandHandler creates a new user create command handler
func NewUserCreateCommandHandler(
	userWriteRepo repositories.UserWriteRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
) *UserCreateCommandHandler {
	return &UserCreateCommandHandler{
		userWriteRepo:  userWriteRepo,
		eventStore:     eventStore,
		eventPublisher: eventPublisher,
	}
}

// Handle handles the create user command
func (h *UserCreateCommandHandler) Handle(ctx context.Context, cmd dto.CreateUserCommand) (*dto.CreateUserCommandResponse, error) {
	// Create user entity with validation
	user, err := entities.NewUser(cmd.Email, cmd.Name)
	if err != nil {
		// Wrap domain validation errors
		return nil, errors.Wrap(err, errors.ErrValidationFailed, "Failed to create user")
	}

	// Check if user already exists
	existingUser, err := h.userWriteRepo.GetByEmail(ctx, cmd.Email)
	if err != nil && !errors.IsAppError(err) {
		return nil, errors.DatabaseError("get user by email", err)
	}
	if existingUser != nil {
		return nil, errors.UserAlreadyExists(cmd.Email)
	}

	// Save to write database (PostgreSQL)
	if err := h.userWriteRepo.Create(ctx, user); err != nil {
		return nil, errors.DatabaseError("create user", err)
	}

	// Create domain event
	userCreatedEvent := &events.UserCreatedEvent{
		UserID:    user.GetID(),
		Email:     user.GetEmail(),
		Name:      user.GetName(),
		CreatedAt: user.CreatedAt,
	}

	// Wrap in Event
	event, err := events.NewEvent("user.created", userCreatedEvent, 1)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrEventStoreFailed, "Failed to create event")
	}

	// Save event to event store
	if err := h.eventStore.SaveEvent(ctx, user.GetID(), event); err != nil {
		return nil, errors.EventStoreError("save event", err)
	}

	// Publish event to Kafka
	if err := h.eventPublisher.PublishEvent(ctx, event); err != nil {
		return nil, errors.EventPublishError(err)
	}

	// Return response
	response := &dto.CreateUserCommandResponse{
		UserID:    user.GetID(),
		Email:     user.GetEmail(),
		Name:      user.GetName(),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return response, nil
}
