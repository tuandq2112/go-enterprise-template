package cmd

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Start gRPC server",
	Long:  `Start the gRPC server that provides gRPC API`,
	Run: func(cmd *cobra.Command, args []string) {
		startGRPCServer()
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)
}

func startGRPCServer() {
	// Use flag port or default to 9090 for gRPC
	grpcPort := "9090"
	if port != "" && port != "8080" {
		grpcPort = port
	}

	// Initialize dependencies using Wire
	grpcServer, err := InitializeGRPCServer()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	log.Printf("Starting gRPC server on port %s", grpcPort)

	// Start server in a goroutine
	go func() {
		if err := grpcServer.GetGRPCServer().Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")

	// Graceful shutdown
	grpcServer.GetGRPCServer().GracefulStop()

	log.Println("gRPC server exited")
}
