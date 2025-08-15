package commands

import (
	"context"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/repositories"
	"go-clean-ddd-es-template/pkg/auth"
	"go-clean-ddd-es-template/pkg/errors"
)

// AuthLoginCommandHandler handles user login
type AuthLoginCommandHandler struct {
	userRepo        repositories.UserRepository
	passwordService *auth.PasswordService
	jwtService      *auth.JWTService
}

// NewAuthLoginCommandHandler creates a new auth login command handler
func NewAuthLoginCommandHandler(
	userRepo repositories.UserRepository,
	passwordService *auth.PasswordService,
	jwtService *auth.JWTService,
) *AuthLoginCommandHandler {
	return &AuthLoginCommandHandler{
		userRepo:        userRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// Handle handles the login command
func (h *AuthLoginCommandHandler) Handle(ctx context.Context, cmd dto.LoginCommand) (*dto.LoginResponse, error) {
	// Get user by email
	user, err := h.userRepo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUserNotFound, "user not found")
	}

	// Check password
	if !h.passwordService.CheckPassword(cmd.Password, user.GetPasswordHash()) {
		return nil, errors.New(errors.ErrUnauthorized, "invalid credentials")
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID.Value(), user.Email.Value(), []string{"user"})
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalServer, "failed to generate token")
	}

	return &dto.LoginResponse{
		UserID: user.ID.Value(),
		Email:  user.Email.Value(),
		Name:   user.Name.Value(),
		Roles:  []string{"user"}, // Default role
		Token:  token,
	}, nil
}
