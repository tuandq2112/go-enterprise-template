package services

import (
	"context"
	"time"

	"go-clean-ddd-es-template/internal/application/commands"
	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/pkg/auth"
)

// AuthService handles authentication and authorization
type AuthService struct {
	registerHandler *commands.AuthRegisterCommandHandler
	loginHandler    *commands.AuthLoginCommandHandler
	jwtService      *auth.JWTService
}

// NewAuthService creates a new auth service
func NewAuthService(
	registerHandler *commands.AuthRegisterCommandHandler,
	loginHandler *commands.AuthLoginCommandHandler,
	jwtService *auth.JWTService,
) *AuthService {
	return &AuthService{
		registerHandler: registerHandler,
		loginHandler:    loginHandler,
		jwtService:      jwtService,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req dto.RegisterCommand) (*dto.RegisterResponse, error) {
	return s.registerHandler.Handle(ctx, req)
}

// Login logs in a user
func (s *AuthService) Login(ctx context.Context, req dto.LoginCommand) (*dto.LoginResponse, error) {
	return s.loginHandler.Handle(ctx, req)
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*dto.ValidateTokenResponse, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return &dto.ValidateTokenResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Roles:  claims.Roles,
	}, nil
}

// RefreshToken refreshes a JWT token
func (s *AuthService) RefreshToken(ctx context.Context, token string) (*dto.RefreshTokenResponse, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Generate new token
	newToken, err := s.jwtService.GenerateToken(claims.UserID, claims.Email, claims.Roles)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		Token: newToken,
	}, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, req dto.ChangePasswordCommand) (*dto.ChangePasswordResponse, error) {
	// TODO: Implement change password handler
	return &dto.ChangePasswordResponse{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}

// GetTokenExpiration returns the token expiration time
func (s *AuthService) GetTokenExpiration() time.Duration {
	return s.jwtService.GetExpiration()
}
