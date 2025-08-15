package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go-clean-ddd-es-template/internal/application/services"
	"go-clean-ddd-es-template/pkg/logger"
)

// AuthInterceptor provides gRPC authentication middleware
type AuthInterceptor struct {
	authService *services.AuthService
	logger      logger.Logger
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(authService *services.AuthService, logger logger.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		authService: authService,
		logger:      logger,
	}
}

// UnaryAuthInterceptor returns a unary interceptor for authentication
func (a *AuthInterceptor) UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for certain methods
		if a.shouldSkipAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		token, err := a.extractToken(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid authorization header: %v", err)
		}

		// Validate token
		claims, err := a.authService.ValidateToken(ctx, token)
		if err != nil {
			a.logger.Error("Token validation failed: %v", err)
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Add user info to context
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_roles", claims.Roles)

		return handler(ctx, req)
	}
}

// StreamAuthInterceptor returns a stream interceptor for authentication
func (a *AuthInterceptor) StreamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip auth for certain methods
		if a.shouldSkipAuth(info.FullMethod) {
			return handler(srv, stream)
		}

		// Extract token from metadata
		token, err := a.extractToken(stream.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "missing or invalid authorization header: %v", err)
		}

		// Validate token
		claims, err := a.authService.ValidateToken(stream.Context(), token)
		if err != nil {
			a.logger.Error("Token validation failed: %v", err)
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Create new context with user info
		ctx := context.WithValue(stream.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_roles", claims.Roles)

		// Create wrapped stream
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

// RoleAuthInterceptor returns a unary interceptor for role-based authorization
func (a *AuthInterceptor) RoleAuthInterceptor(requiredRoles ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get user roles from context
		userRoles, ok := ctx.Value("user_roles").([]string)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
		}

		// Check if user has required roles
		hasRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			a.logger.Error("User does not have required roles - user_roles: %v, required_roles: %v", userRoles, requiredRoles)
			return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
		}

		return handler(ctx, req)
	}
}

// shouldSkipAuth checks if authentication should be skipped for a method
func (a *AuthInterceptor) shouldSkipAuth(method string) bool {
	// Skip auth for these methods
	skipMethods := []string{
		"/auth.AuthService/Register",
		"/auth.AuthService/Login",
		"/grpc.health.v1.Health/Check",
		"/grpc.health.v1.Health/Watch",
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
	}

	for _, skipMethod := range skipMethods {
		if method == skipMethod {
			return true
		}
	}

	// Skip reflection methods
	if strings.Contains(method, "grpc.reflection") {
		return true
	}

	return false
}

// extractToken extracts JWT token from gRPC metadata
func (a *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata not found")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization header not found")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", status.Errorf(codes.Unauthenticated, "token not found")
	}

	return token, nil
}

// wrappedServerStream wraps grpc.ServerStream to override context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
