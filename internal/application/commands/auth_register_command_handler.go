package commands

import (
	"context"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/domain/repositories"
	"go-clean-ddd-es-template/pkg/auth"
	"go-clean-ddd-es-template/pkg/errors"
)

// AuthRegisterCommandHandler handles user registration
type AuthRegisterCommandHandler struct {
	userRepo        repositories.UserRepository
	eventStore      repositories.EventStore
	eventPublisher  repositories.EventPublisher
	passwordService *auth.PasswordService
	jwtService      *auth.JWTService
}

// NewAuthRegisterCommandHandler creates a new auth register command handler
func NewAuthRegisterCommandHandler(
	userRepo repositories.UserRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
	passwordService *auth.PasswordService,
	jwtService *auth.JWTService,
) *AuthRegisterCommandHandler {
	return &AuthRegisterCommandHandler{
		userRepo:        userRepo,
		eventStore:      eventStore,
		eventPublisher:  eventPublisher,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// Handle handles the register command
func (h *AuthRegisterCommandHandler) Handle(ctx context.Context, cmd dto.RegisterCommand) (*dto.RegisterResponse, error) {
	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(ctx, cmd.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New(errors.ErrUserAlreadyExists, "user already exists")
	}

	// Validate password
	if err := h.passwordService.ValidatePassword(cmd.Password); err != nil {
		return nil, errors.Wrap(err, errors.ErrValidationFailed, "invalid password")
	}

	// Hash password
	hashedPassword, err := h.passwordService.HashPassword(cmd.Password)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalServer, "failed to hash password")
	}

	// Create user
	user, err := entities.NewUser(cmd.Email, cmd.Name)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrValidationFailed, "failed to create user")
	}

	// Set password hash
	user.SetPasswordHash(hashedPassword)

	// Create user created event
	userCreatedEvent := &events.UserCreatedEvent{
		UserID:    user.ID.Value(),
		Email:     user.Email.Value(),
		Name:      user.Name.Value(),
		CreatedAt: user.CreatedAt,
	}

	// Create domain event
	event, err := events.NewEvent("user.created", userCreatedEvent, 1)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrEventStoreFailed, "failed to create event")
	}

	// Save event to event store
	if err := h.eventStore.SaveEvent(ctx, user.ID.Value(), event); err != nil {
		return nil, errors.Wrap(err, errors.ErrEventStoreFailed, "failed to save event")
	}

	// Publish event
	if err := h.eventPublisher.PublishEvent(ctx, event); err != nil {
		return nil, errors.Wrap(err, errors.ErrEventPublishFailed, "failed to publish event")
	}

	// Save user to write database
	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalServer, "failed to save user")
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID.Value(), user.Email.Value(), []string{"user"})
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalServer, "failed to generate token")
	}

	return &dto.RegisterResponse{
		UserID: user.ID.Value(),
		Email:  user.Email.Value(),
		Name:   user.Name.Value(),
		Token:  token,
	}, nil
}
