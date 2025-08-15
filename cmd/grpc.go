package cmd

import (
	"context"
	"os"

	"go-clean-ddd-es-template/internal/infrastructure/grpc"

	"github.com/spf13/cobra"
)

var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Start the gRPC server with HTTP gateway",
	Long:  `Start the gRPC server with HTTP gateway and Swagger UI`,
	Run: func(cmd *cobra.Command, args []string) {
		startGRPCServer()
	},
}

func startGRPCServer() {
	// Use flag port or default to 9091 for gRPC (9090 is used by Prometheus)
	grpcPort := "9091"
	gatewayPort := "8080"

	if port != "" && port != "8080" {
		grpcPort = port
	}

	// Initialize dependencies using Wire
	grpcServer, err := InitializeGRPCServer()
	if err != nil {
		// Use standard log for initialization errors since logger might not be available yet
		os.Stderr.WriteString("Failed to initialize dependencies: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Initialize event consumer
	eventConsumer, err := InitializeEventConsumer()
	if err != nil {
		os.Stderr.WriteString("Failed to initialize event consumer: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Get logger from the server
	logger := grpcServer.GetLogger()
	if logger == nil {
		// Fallback to standard log if logger is not available
		os.Stderr.WriteString("Logger not available, using standard logging\n")
	}

	// Create HTTP server instance
	httpServer := grpc.NewHTTPServer(grpcServer, logger)

	if logger != nil {
		logger.Info("Starting gRPC server on port %s and HTTP gateway on port %s", grpcPort, gatewayPort)
		logger.Info("Starting event consumer...")
	} else {
		os.Stdout.WriteString("Starting gRPC server on port " + grpcPort + " and HTTP gateway on port " + gatewayPort + "\n")
		os.Stdout.WriteString("Starting event consumer...\n")
	}

	// Start event consumer in background
	ctx := context.Background()
	go func() {
		if err := eventConsumer.Start(ctx); err != nil {
			if logger != nil {
				logger.Error("Failed to start event consumer: %v", err)
			} else {
				os.Stderr.WriteString("Failed to start event consumer: " + err.Error() + "\n")
			}
		}
	}()

	// Start HTTP server
	if err := httpServer.Start(grpcPort, gatewayPort); err != nil {
		if logger != nil {
			logger.Error("Failed to start server: %v", err)
		} else {
			os.Stderr.WriteString("Failed to start server: " + err.Error() + "\n")
		}
		os.Exit(1)
	}
}
