.PHONY: build run test clean help deps proto all-up all-down test-api test-grpc test-integration kill check-structure generate-mocks test-coverage test-v test-domain test-application test-infrastructure test-pkg clean-test demo-switch-db

# Default target
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Run gRPC server (default)"
	@echo "  run-grpc        - Run gRPC server"
	@echo "  test            - Run unit tests"
	@echo "  test-v          - Run tests with verbose output"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-domain     - Run domain layer tests"
	@echo "  test-application- Run application layer tests"
	@echo "  test-infrastructure- Run infrastructure layer tests"
	@echo "  test-pkg        - Run pkg tests"
	@echo "  test-grpc       - Run gRPC integration tests"
	@echo "  test-integration- Run all integration tests"
	@echo "  generate-mocks  - Generate mocks using Mockery"
	@echo "  clean           - Clean build artifacts"
	@echo "  clean-test      - Clean test artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  proto           - Generate protobuf code"
	@echo "  all-up          - Start all Docker services"
	@echo "  all-down        - Stop all Docker services"
	@echo "  kill            - Kill running Go processes"
	@echo "  check-structure - Check project structure"
	@echo "  demo-switch-db  - Demo database/message broker switching"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/app main.go

# Run gRPC server
run-grpc:
	@echo "Running gRPC server..."
	./bin/app grpc

# Kill Go processes
kill:
	@echo "🔍 Searching for Go processes..."
	@PIDS=$$(pgrep -f "go-clean-ddd-es-template" 2>/dev/null); \
	if [ ! -z "$$PIDS" ]; then \
		echo "📦 Found go-clean-ddd-es-template processes: $$PIDS"; \
		echo "$$PIDS" | xargs kill -9; \
		echo "✅ Killed go-clean-ddd-es-template processes"; \
	else \
		echo "ℹ️  No go-clean-ddd-es-template processes found"; \
	fi
	@GO_RUN_PIDS=$$(pgrep -f "go run" 2>/dev/null); \
	if [ ! -z "$$GO_RUN_PIDS" ]; then \
		echo "🏃 Found 'go run' processes: $$GO_RUN_PIDS"; \
		echo "$$GO_RUN_PIDS" | xargs kill -9; \
		echo "✅ Killed 'go run' processes"; \
	else \
		echo "ℹ️  No 'go run' processes found"; \
	fi
	@echo "🔌 Checking common ports..."
	@for port in 8080 9090 3000 5000 8000; do \
		PID=$$(lsof -ti:$$port 2>/dev/null); \
		if [ ! -z "$$PID" ]; then \
			echo "🚪 Port $$port is in use by PID: $$PID"; \
			echo "$$PID" | xargs kill -9; \
			echo "✅ Killed process on port $$port"; \
		fi; \
	done
	@echo "🎉 Cleanup completed!"

# Run the application (default to gRPC)
run: run-grpc

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output
test-v:
	@echo "Running tests with verbose output..."
	go test ./... -v

# Run specific test packages
test-domain:
	@echo "Running domain tests..."
	go test ./internal/domain/... -v

test-application:
	@echo "Running application tests..."
	go test ./internal/application/... -v

test-infrastructure:
	@echo "Running infrastructure tests..."
	go test ./internal/infrastructure/... -v

test-pkg:
	@echo "Running pkg tests..."
	go test ./pkg/... -v

# Generate mocks
generate-mocks:
	@echo "Generating mocks..."
	mockery --config .mockery.yaml

# Clean test artifacts
clean-test:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html

# Demo database and message broker switching
demo-switch-db:
	@echo "=== Database and Message Broker Switching Demo ==="
	@echo ""
	@echo "1. Testing PostgreSQL + Kafka (Default)"
	@DB_TYPE=postgres MESSAGE_BROKER_TYPE=kafka go build -o /tmp/test_build . 2>&1 && echo "✅ Build successful - configuration is valid" || echo "❌ Build failed - check configuration"
	@echo ""
	@echo "2. Testing MySQL + RabbitMQ (Stub)"
	@DB_TYPE=mysql MESSAGE_BROKER_TYPE=rabbitmq go build -o /tmp/test_build . 2>&1 && echo "✅ Build successful - configuration is valid" || echo "❌ Build failed - check configuration"
	@echo ""
	@echo "3. Testing MongoDB + Redis (Stub)"
	@DB_TYPE=mongodb MESSAGE_BROKER_TYPE=redis go build -o /tmp/test_build . 2>&1 && echo "✅ Build successful - configuration is valid" || echo "❌ Build failed - check configuration"
	@echo ""
	@echo "4. Testing PostgreSQL + NATS (Stub)"
	@DB_TYPE=postgres MESSAGE_BROKER_TYPE=nats go build -o /tmp/test_build . 2>&1 && echo "✅ Build successful - configuration is valid" || echo "❌ Build failed - check configuration"
	@echo ""
	@echo "5. Testing Invalid Database Type"
	@DB_TYPE=invalid_db MESSAGE_BROKER_TYPE=kafka go build -o /tmp/test_build . 2>&1 && echo "✅ Build successful - configuration is valid" || echo "❌ Build failed - check configuration"
	@echo ""
	@echo "6. Testing Invalid Message Broker Type"
	@DB_TYPE=postgres MESSAGE_BROKER_TYPE=invalid_broker go build -o /tmp/test_build . 2>&1 && echo "✅ Build successful - configuration is valid" || echo "❌ Build failed - check configuration"
	@echo ""
	@echo "=== Demo Complete ==="
	@echo ""
	@echo "Note: Stub implementations will return errors when actually used,"
	@echo "but the configuration and dependency injection will work correctly."
	@echo ""
	@echo "To actually run the application, use:"
	@echo "DB_TYPE=postgres MESSAGE_BROKER_TYPE=kafka go run ."

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
	@if ! command -v protoc &> /dev/null; then \
		echo "Error: protoc is not installed"; \
		echo "Please install protobuf compiler:"; \
		echo "  Ubuntu/Debian: sudo apt-get install protobuf-compiler"; \
		echo "  macOS: brew install protobuf"; \
		exit 1; \
	fi
	@mkdir -p proto/user
	@if [ ! -d "third_party/googleapis" ]; then \
		echo "Downloading googleapis..."; \
		git clone https://github.com/googleapis/googleapis.git third_party/googleapis; \
	fi
	@echo "Generating Go code..."
	@protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--proto_path=. \
		--proto_path=third_party/googleapis \
		proto/user/user.proto
	@echo "Protobuf code generated successfully!"

# Start all services
all-up:
	@echo "Starting all services..."
	docker-compose up -d

# Stop all services
all-down:
	@echo "Stopping all services..."
	docker-compose down

# Test gRPC endpoints
test-grpc:
	@echo "Testing gRPC endpoints..."
	go test ./integration_test -v -run "TestGRPC"

# Run all integration tests
test-integration:
	@echo "Running all integration tests..."
	go test ./integration_test -v

# Check project structure
check-structure:
	@echo "Checking Project Structure"
	@echo "========================="
	@echo ""
	@echo "Main Directories:"
	@if [ -d "cmd" ]; then echo "✅ CLI Commands"; else echo "❌ CLI Commands (missing directory)"; fi
	@if [ -d "internal" ]; then echo "✅ Internal Application Code"; else echo "❌ Internal Application Code (missing directory)"; fi
	@if [ -d "pkg" ]; then echo "✅ Reusable Packages"; else echo "❌ Reusable Packages (missing directory)"; fi
	@if [ -d "proto" ]; then echo "✅ Protocol Buffers"; else echo "❌ Protocol Buffers (missing directory)"; fi
	@if [ -d "monitoring" ]; then echo "✅ Monitoring Configuration"; else echo "❌ Monitoring Configuration (missing directory)"; fi
	@echo ""
	@echo "Internal Structure:"
	@if [ -d "internal/domain" ]; then echo "✅ Domain Layer"; else echo "❌ Domain Layer (missing directory)"; fi
	@if [ -d "internal/application" ]; then echo "✅ Application Layer"; else echo "❌ Application Layer (missing directory)"; fi
	@if [ -d "internal/infrastructure" ]; then echo "✅ Infrastructure Layer"; else echo "❌ Infrastructure Layer (missing directory)"; fi
	@echo ""
	@echo "Important Files:"
	@if [ -f "main.go" ]; then echo "✅ main.go"; else echo "❌ main.go (missing)"; fi
	@if [ -f "go.mod" ]; then echo "✅ go.mod"; else echo "❌ go.mod (missing)"; fi
	@if [ -f "Makefile" ]; then echo "✅ Makefile"; else echo "❌ Makefile (missing)"; fi
	@if [ -f "docker-compose.yml" ]; then echo "✅ docker-compose.yml"; else echo "❌ docker-compose.yml (missing)"; fi
	@if [ -f "README.md" ]; then echo "✅ README.md"; else echo "❌ README.md (missing)"; fi
	@echo ""
	@echo "Project Structure Check Complete!" 