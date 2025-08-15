package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Global variables for flags
var port string

var rootCmd = &cobra.Command{
	Use:   "go-clean-ddd-es-template",
	Short: "A Go application with Clean Architecture, DDD, and Event Sourcing",
	Long: `A Go application template implementing:
- Clean Architecture principles
- Domain-Driven Design (DDD)
- Event Sourcing
- Kafka integration
- PostgreSQL database
- Protocol Buffers and gRPC
- gRPC Gateway for REST API`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")

	// Add commands
	rootCmd.AddCommand(grpcCmd)
}
