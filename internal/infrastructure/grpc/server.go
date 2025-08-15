package grpc

import (
	"context"
	"net"
	"net/http"

	"go-clean-ddd-es-template/pkg/logger"
)

// HTTPServer represents the HTTP server that serves both gRPC and HTTP gateway
type HTTPServer struct {
	grpcServer *GRPCServer
	logger     logger.Logger
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(grpcServer *GRPCServer, logger logger.Logger) *HTTPServer {
	return &HTTPServer{
		grpcServer: grpcServer,
		logger:     logger,
	}
}

// Start starts the gRPC server and HTTP gateway
func (s *HTTPServer) Start(grpcPort, gatewayPort string) error {
	// Start gRPC server in background
	go func() {
		s.logger.Info("Starting gRPC server on port: %s", grpcPort)
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			s.logger.Fatal("Failed to listen for gRPC: %v", err)
		}

		if err := s.grpcServer.GetGRPCServer().Serve(lis); err != nil {
			s.logger.Fatal("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP gateway (main server)
	s.logger.Info("Starting HTTP gateway on port: %s", gatewayPort)
	mux := http.NewServeMux()

	// Add swagger handlers
	swaggerHandler := NewSwaggerHandler("docs/swagger.json")
	mux.HandleFunc("/docs", swaggerHandler.ServeSwaggerIndex)
	mux.HandleFunc("/swagger", swaggerHandler.ServeSwaggerUI)
	mux.HandleFunc("/swagger/", swaggerHandler.ServeSwaggerUI)
	mux.HandleFunc("/swagger.json", swaggerHandler.ServeSwaggerJSON)

	// Add gRPC gateway handler
	mux.Handle("/", s.grpcServer)

	server := &http.Server{
		Addr:    ":" + gatewayPort,
		Handler: mux,
	}

	return server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server...")

	// Graceful shutdown of gRPC server
	s.grpcServer.GetGRPCServer().GracefulStop()

	s.logger.Info("HTTP server stopped successfully")
	return nil
}

// GetGRPCServer returns the underlying gRPC server
func (s *HTTPServer) GetGRPCServer() *GRPCServer {
	return s.grpcServer
}
