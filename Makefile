.PHONY: help build run test clean deps proto migrate-up migrate-down generate-keys all-up all-down

# Default target
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Run gRPC server"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  proto           - Generate protobuf code"
	@echo "  migrate-up      - Run database migrations"
	@echo "  migrate-down    - Rollback migrations"
	@echo "  generate-keys   - Generate RSA keys"
	@echo "  all-up          - Start Docker services"
	@echo "  all-down        - Stop Docker services"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/app main.go

# Run gRPC server
run:
	@echo "Running gRPC server..."
	./bin/app grpc

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	buf generate
	@echo "Protobuf code generated successfully!"

# Database migrations
migrate-up:
	@echo "Running migrations..."
	@if [ ! -f "bin/app" ]; then make build; fi
	./bin/app migrate up

migrate-down:
	@echo "Rolling back migrations..."
	@if [ ! -f "bin/app" ]; then make build; fi
	./bin/app migrate down

# Generate RSA keys
generate-keys:
	@echo "Generating RSA keys..."
	@if [ ! -f "bin/app" ]; then make build; fi
	./bin/app generate-keys

# Docker services
all-up:
	@echo "Starting Docker services..."
	docker-compose up -d

all-down:
	@echo "Stopping Docker services..."
	docker-compose down 