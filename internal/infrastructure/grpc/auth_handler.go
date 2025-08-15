package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/application/services"
	"go-clean-ddd-es-template/pkg/logger"
	"go-clean-ddd-es-template/proto/auth"
)

// AuthHandler handles gRPC auth requests
type AuthHandler struct {
	auth.UnimplementedAuthServiceServer
	authService *services.AuthService
	logger      logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	h.logger.Info("Handling register request for email: %s, name: %s", req.Email, req.Name)

	// Convert gRPC request to service request
	serviceReq := dto.RegisterCommand{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}

	// Call auth service
	resp, err := h.authService.Register(ctx, serviceReq)
	if err != nil {
		h.logger.Error("Failed to register user: %v, email: %s", err, req.Email)
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	// Convert service response to gRPC response
	return &auth.RegisterResponse{
		UserId: resp.UserID,
		Email:  resp.Email,
		Name:   resp.Name,
		Token:  resp.Token,
	}, nil
}

// Login handles user login
func (h *AuthHandler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	h.logger.Info("Handling login request for email: %s", req.Email)

	// Convert gRPC request to service request
	serviceReq := dto.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	// Call auth service
	resp, err := h.authService.Login(ctx, serviceReq)
	if err != nil {
		h.logger.Error("Failed to login user: %v, email: %s", err, req.Email)
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
	}

	// Convert service response to gRPC response
	return &auth.LoginResponse{
		UserId: resp.UserID,
		Email:  resp.Email,
		Name:   resp.Name,
		Roles:  resp.Roles,
		Token:  resp.Token,
	}, nil
}

// ValidateToken validates JWT token
func (h *AuthHandler) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	h.logger.Info("Handling validate token request")

	// Call auth service
	claims, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		h.logger.Error("Failed to validate token: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Get token expiration
	expiration := h.authService.GetTokenExpiration()

	// Convert service response to gRPC response
	return &auth.ValidateTokenResponse{
		UserId:    claims.UserID,
		Email:     claims.Email,
		Roles:     claims.Roles,
		ExpiresAt: timestamppb.New(time.Now().Add(expiration)),
	}, nil
}

// RefreshToken refreshes JWT token
func (h *AuthHandler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	h.logger.Info("Handling refresh token request")

	// Call auth service
	resp, err := h.authService.RefreshToken(ctx, req.Token)
	if err != nil {
		h.logger.Error("Failed to refresh token: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "failed to refresh token: %v", err)
	}

	// Get new token expiration
	expiration := h.authService.GetTokenExpiration()

	// Convert service response to gRPC response
	return &auth.RefreshTokenResponse{
		Token:     resp.Token,
		ExpiresAt: timestamppb.New(time.Now().Add(expiration)),
	}, nil
}

// ChangePassword changes user password
func (h *AuthHandler) ChangePassword(ctx context.Context, req *auth.ChangePasswordRequest) (*auth.ChangePasswordResponse, error) {
	h.logger.Info("Handling change password request")

	// Get user ID from context (should be set by auth middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error("User ID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	// Convert gRPC request to service request
	serviceReq := dto.ChangePasswordCommand{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	// Call auth service
	resp, err := h.authService.ChangePassword(ctx, serviceReq)
	if err != nil {
		h.logger.Error("Failed to change password: %v, user_id: %s", err, userID)
		return nil, status.Errorf(codes.Internal, "failed to change password: %v", err)
	}

	// Convert service response to gRPC response
	return &auth.ChangePasswordResponse{
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}
