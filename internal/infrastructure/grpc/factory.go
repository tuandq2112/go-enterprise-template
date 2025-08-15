package grpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"go-clean-ddd-es-template/internal/application/services"
	"go-clean-ddd-es-template/pkg/logger"
	"go-clean-ddd-es-template/pkg/middleware"
	"go-clean-ddd-es-template/pkg/tracing"
	"go-clean-ddd-es-template/proto/auth"
	"go-clean-ddd-es-template/proto/user"
)

// GRPCServer represents the gRPC server with gateway
type GRPCServer struct {
	grpcServer  *grpc.Server
	gatewayMux  *runtime.ServeMux
	userService *services.UserService
	authService *services.AuthService
	tracer      *tracing.Tracer
	logger      logger.Logger
}

// GetGRPCServer returns the gRPC server
func (s *GRPCServer) GetGRPCServer() *grpc.Server {
	return s.grpcServer
}

// GetGatewayMux returns the gateway mux
func (s *GRPCServer) GetGatewayMux() *runtime.ServeMux {
	return s.gatewayMux
}

// GetLogger returns the logger
func (s *GRPCServer) GetLogger() logger.Logger {
	return s.logger
}

// ServeHTTP implements http.Handler for the gateway
func (s *GRPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.gatewayMux.ServeHTTP(w, r)
}

// NewGRPCServer creates a new gRPC server with gateway
func NewGRPCServer(userService *services.UserService, authService *services.AuthService, tracer *tracing.Tracer, logger logger.Logger) *GRPCServer {
	// Create validation middleware
	validationConfig := middleware.DefaultValidationConfig()
	// Adjust config for gRPC (higher limits, different rate limiting)
	validationConfig.MaxRequestSize = 50 * 1024 * 1024 // 50MB for gRPC
	validationConfig.MaxHeaderSize = 5 * 1024 * 1024   // 5MB for gRPC headers
	validationConfig.RateLimitRequests = 1000          // Higher rate limit for gRPC
	validationConfig.RateLimitWindow = 60 * 60         // 1 hour window

	validationMiddleware := middleware.NewValidationMiddleware(validationConfig, logger)

	// Create auth interceptor
	authInterceptor := middleware.NewAuthInterceptor(authService, logger)

	// Create gRPC server with interceptors
	var opts []grpc.ServerOption
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// Add tracing interceptors
	if tracer != nil {
		unaryInterceptors = append(unaryInterceptors, middleware.GRPCTracingInterceptor(tracer))
		streamInterceptors = append(streamInterceptors, middleware.GRPCStreamTracingInterceptor(tracer))
	}

	// Add validation interceptors
	unaryInterceptors = append(unaryInterceptors, middleware.GRPCValidationInterceptor(validationMiddleware))
	streamInterceptors = append(streamInterceptors, middleware.GRPCStreamValidationInterceptor(validationMiddleware))

	// Add rate limiting interceptors
	unaryInterceptors = append(unaryInterceptors, middleware.GRPCRateLimitInterceptor(validationMiddleware))
	streamInterceptors = append(streamInterceptors, middleware.GRPCStreamRateLimitInterceptor(validationMiddleware))

	// Add auth interceptors
	unaryInterceptors = append(unaryInterceptors, authInterceptor.UnaryAuthInterceptor())
	streamInterceptors = append(streamInterceptors, authInterceptor.StreamAuthInterceptor())

	// Chain all interceptors
	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	grpcServer := grpc.NewServer(opts...)

	// Create user gRPC server
	userGRPCServer := NewUserGRPCServer(userService, tracer)

	// Create auth gRPC server
	authGRPCServer := NewAuthHandler(authService, logger)

	// Register services
	user.RegisterUserServiceServer(grpcServer, userGRPCServer)
	auth.RegisterAuthServiceServer(grpcServer, authGRPCServer)

	// Register reflection service on gRPC server
	reflection.Register(grpcServer)

	// Create gRPC Gateway mux with validation middleware
	gatewayMux := runtime.NewServeMux()

	// Register gRPC Gateway handlers
	gatewayOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register user service gateway
	if err := user.RegisterUserServiceHandlerFromEndpoint(
		context.Background(),
		gatewayMux,
		"localhost:9091", // gRPC server address
		gatewayOpts,
	); err != nil {
		panic(fmt.Sprintf("failed to register user gateway: %v", err))
	}

	// Register auth service gateway
	if err := auth.RegisterAuthServiceHandlerFromEndpoint(
		context.Background(),
		gatewayMux,
		"localhost:9091", // gRPC server address
		gatewayOpts,
	); err != nil {
		panic(fmt.Sprintf("failed to register auth gateway: %v", err))
	}

	return &GRPCServer{
		grpcServer:  grpcServer,
		gatewayMux:  gatewayMux,
		userService: userService,
		authService: authService,
		tracer:      tracer,
		logger:      logger,
	}
}
