# Go Clean DDD ES Template

A Go project template implementing **Clean Architecture**, **Domain-Driven Design (DDD)**, **Event Sourcing**, and **CQRS** principles.

## 🚀 Features

- **Clean Architecture**: Proper separation of concerns
- **Domain-Driven Design**: Rich domain models
- **Event Sourcing**: All changes captured as events
- **CQRS**: Separate read and write models
- **gRPC & REST**: Dual API support
- **Authentication**: JWT with RSA signing
- **Database Migrations**: Multi-database support
- **Structured Logging**: Zap logger
- **Input Validation**: Security middleware
- **Monitoring**: Prometheus, Grafana integration

## 📁 Project Structure

```
go-clean-ddd-es-template/
├── cmd/                    # CLI commands
├── internal/               # Application code
│   ├── domain/            # Domain layer
│   ├── application/       # Application layer
│   └── infrastructure/    # Infrastructure layer
├── pkg/                   # Reusable packages
├── proto/                 # Protocol Buffers
├── migrations/            # Database migrations
├── scripts/               # Utility scripts
└── monitoring/            # Monitoring config
```

## 🛠️ Prerequisites

- **Go 1.21+**
- **Docker & Docker Compose**
- **Buf** (for protobuf generation)

## 🚀 Step-by-Step Setup

### Step 1: Clone and Setup Project

```bash
# 1. Clone the repository
git clone <repository-url>
cd go-clean-ddd-es-template

# 2. Copy environment file
cp env.example .env

# 3. Install Go dependencies
make deps
```

### Step 2: Start Database

```bash
# 4. Start PostgreSQL database
make all-up

# Or if you only want PostgreSQL:
docker-compose up -d postgres

# 5. Verify database is running
docker ps
# You should see postgres container running
```

### Step 3: Setup Database Schema

```bash
# 6. Build the application first
make build

# 7. Run database migrations
make migrate-up

# Expected output:
# Running migrations...
# Connected to PostgreSQL database: clean_ddd_write_db
# Connected to PostgreSQL database: clean_ddd_event_db
# Running write database migrations...
# Write database migrations completed
# Running event database migrations...
# Event database migrations completed
```

### Step 4: Generate Authentication Keys

```bash
# 8. Generate RSA keys for JWT authentication
make generate-keys

# Expected output:
# Generating RSA keys...
# RSA key pair generated successfully!
# Private key: keys/private.pem
# Public key: keys/public.pem
```

### Step 5: Generate Protocol Buffer Code

```bash
# 9. Generate gRPC and protobuf code
make proto

# Expected output:
# Generating protobuf code...
# buf generate
# Protobuf code generated successfully!
```

### Step 6: Build and Run Application

```bash
# 10. Build the application
make build

# 11. Run the gRPC server with HTTP gateway
make run

# Expected output:
# Running gRPC server...
# Starting gRPC server on port 9090 and HTTP gateway on port 8080
```

### Step 7: Verify Everything is Working

```bash
# 12. Check if server is running
curl http://localhost:8080/docs

# 13. Test API documentation
open http://localhost:8080/docs
# or visit in browser: http://localhost:8080/docs
```

## 📚 API Testing

### Test User Management

```bash
# Create a new user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe"
  }'

# Expected response:
# {
#   "id": "uuid-here",
#   "email": "john@example.com",
#   "name": "John Doe",
#   "created_at": "2025-08-15T..."
# }

# List all users
curl http://localhost:8080/api/v1/users

# Get user by ID (replace {id} with actual user ID)
curl http://localhost:8080/api/v1/users/{id}
```

### Test Authentication

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "name": "Jane Doe",
    "password": "password123"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "password123"
  }'

# Expected response:
# {
#   "token": "jwt-token-here",
#   "user": {
#     "id": "uuid-here",
#     "email": "jane@example.com",
#     "name": "Jane Doe"
#   }
# }
```

## 🔧 Development Commands

```bash
# Build and run
make build          # Build application
make run            # Run gRPC server

# Testing
make test           # Run all tests

# Database operations
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations

# Code generation
make proto          # Generate protobuf code

# Docker operations
make all-up         # Start all services
make all-down       # Stop all services

# Show all available commands
make help
```

## ⚙️ Configuration

Copy `env.example` to `.env` and configure:

```bash
# Server ports
PORT=8080           # HTTP gateway port
GRPC_PORT=9091      # gRPC server port

# Database configuration
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=clean_ddd_write_db

# Logging
LOG_LEVEL=debug     # debug, info, warn, error
LOG_FORMAT=json     # json or text

# Authentication
AUTH_PRIVATE_KEY_PATH=keys/private.pem
AUTH_PUBLIC_KEY_PATH=keys/public.pem
AUTH_TOKEN_EXPIRY=24h
```

## 🏗️ Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────┐
│        Infrastructure          │
│  gRPC | Database | Kafka       │
└─────────────────────────────────┘
              │
┌─────────────────────────────────┐
│         Application            │
│  Commands | Queries | Services │
└─────────────────────────────────┘
              │
┌─────────────────────────────────┐
│           Domain               │
│  Entities | Events | Repos     │
└─────────────────────────────────┘
```

### Event Sourcing Flow

```
Command → Domain Entity → Event → Event Store → Read Model
                ↓
            Message Broker
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test packages
go test ./internal/domain/...
go test ./internal/application/...
go test ./internal/infrastructure/...
```

## 🚀 Deployment

### Docker

```bash
# Build image
docker build -t go-clean-ddd-es-template .

# Run with compose
docker-compose up -d
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

MIT License 