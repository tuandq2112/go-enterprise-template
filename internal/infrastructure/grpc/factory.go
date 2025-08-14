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
	"go-clean-ddd-es-template/pkg/middleware"
	"go-clean-ddd-es-template/pkg/tracing"
	"go-clean-ddd-es-template/proto/user"
)

// GRPCServer represents the gRPC server with gateway
type GRPCServer struct {
	grpcServer  *grpc.Server
	gatewayMux  *runtime.ServeMux
	userService *services.UserService
	tracer      *tracing.Tracer
}

// GetGRPCServer returns the gRPC server
func (s *GRPCServer) GetGRPCServer() *grpc.Server {
	return s.grpcServer
}

// GetGatewayMux returns the gateway mux
func (s *GRPCServer) GetGatewayMux() *runtime.ServeMux {
	return s.gatewayMux
}

// ServeHTTP implements http.Handler for the gateway
func (s *GRPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.gatewayMux.ServeHTTP(w, r)
}

// NewGRPCServer creates a new gRPC server with gateway
func NewGRPCServer(userService *services.UserService, tracer *tracing.Tracer) *GRPCServer {
	// Create gRPC server with interceptors
	var opts []grpc.ServerOption

	if tracer != nil {
		opts = append(opts, grpc.UnaryInterceptor(middleware.GRPCTracingInterceptor(tracer)))
		opts = append(opts, grpc.StreamInterceptor(middleware.GRPCStreamTracingInterceptor(tracer)))
	}

	grpcServer := grpc.NewServer(opts...)

	// Create user gRPC server
	userGRPCServer := NewUserGRPCServer(userService, tracer)

	// Register user service
	user.RegisterUserServiceServer(grpcServer, userGRPCServer)

	// Register reflection service on gRPC server
	reflection.Register(grpcServer)

	// Create gRPC Gateway mux
	gatewayMux := runtime.NewServeMux()

	// Register gRPC Gateway handlers
	gatewayOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register user service gateway
	if err := user.RegisterUserServiceHandlerFromEndpoint(
		context.Background(),
		gatewayMux,
		"localhost:9090", // gRPC server address
		gatewayOpts,
	); err != nil {
		panic(fmt.Sprintf("failed to register gateway: %v", err))
	}

	return &GRPCServer{
		grpcServer:  grpcServer,
		gatewayMux:  gatewayMux,
		userService: userService,
		tracer:      tracer,
	}
}
