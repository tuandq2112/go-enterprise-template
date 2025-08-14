# Go Clean DDD ES Template

A Go project template implementing Clean Architecture, Domain-Driven Design (DDD), and Event Sourcing principles with Kafka integration.

## Project Structure

```
go-clean-ddd-es-template/
├── main.go                                    # Application entry point
├── go.mod                                     # Go module file
├── README.md                                  # This file
├── Makefile                                   # Development tasks
├── .gitignore                                 # Git ignore rules
└── internal/                                  # Private application code
    ├── domain/                               # Domain layer (business logic)
    │   ├── entities/                         # Domain entities (User)
    │   ├── repositories/                     # Repository interfaces
    │   ├── events/                           # Domain events
    │   └── services/                         # Domain services
    ├── application/                          # Application layer (use cases)
    │   ├── usecases/                         # Use cases (UserUseCase)
    │   └── interfaces/                       # Application interfaces
    └── infrastructure/                       # Infrastructure layer
        ├── config/                           # Configuration management
        ├── http/                             # HTTP server and handlers
        │   ├── server/                       # HTTP server setup
        │   └── handlers/                     # HTTP handlers
        ├── kafka/                            # Kafka integration
        ├── database/                         # Database connections
        └── repositories/                     # Repository implementations
```

## Architecture

This project follows Clean Architecture principles with the following layers:

1. **Domain Layer**: Contains business entities, repository interfaces, domain services, and domain events
2. **Application Layer**: Contains use cases and application services
3. **Infrastructure Layer**: Contains external concerns like HTTP, database, Kafka, and configuration

## Features

- **Clean Architecture**: Proper separation of concerns
- **Domain-Driven Design**: Rich domain models with business logic
- **Event Sourcing**: All domain changes are captured as events
- **Kafka Integration**: Event streaming with Apache Kafka
- **RESTful API**: Complete CRUD operations for users
- **In-Memory Storage**: For development and testing
- **UUID Generation**: Unique identifiers for entities

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Apache Kafka (optional, for event streaming)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-clean-ddd-es-template
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run main.go
```

The server will start on port 8080 by default.

### Environment Variables

You can configure the application using the following environment variables:

- `PORT`: Server port (default: 8080)
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: password)
- `DB_NAME`: Database name (default: clean_ddd_db)
- `KAFKA_BROKERS`: Kafka broker addresses (default: localhost:9092)

## API Endpoints

### Health Check
- `GET /health` - Health check endpoint

### API Root
- `GET /api/v1/` - API root endpoint

### User Management
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - List all users
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### Example API Usage

#### Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com", "name": "John Doe"}'
```

#### Get User
```bash
curl http://localhost:8080/api/v1/users/{user-id}
```

#### Update User
```bash
curl -X PUT http://localhost:8080/api/v1/users/{user-id} \
  -H "Content-Type: application/json" \
  -d '{"name": "John Smith", "email": "john.smith@example.com"}'
```

#### Delete User
```bash
curl -X DELETE http://localhost:8080/api/v1/users/{user-id}
```

## Event Sourcing

The application implements Event Sourcing where all domain changes are captured as events:

- `user.created` - When a user is created
- `user.updated` - When a user is updated
- `user.deleted` - When a user is deleted

Events are stored in an event store and can be published to Kafka for event streaming.

## Kafka Integration

The application can publish domain events to Kafka topics:

- Topic: `user-events` - Contains all user-related events
- Event Types: `user.created`, `user.updated`, `user.deleted`

To enable Kafka:
1. Start a Kafka broker
2. Set the `KAFKA_BROKERS` environment variable
3. The application will automatically publish events to Kafka

## Protocol Buffers and gRPC

The application uses Protocol Buffers for API definition and supports both gRPC and REST APIs:

### Protocol Buffers
- Service definitions in `proto/user/user.proto`
- Code generation using Buf
- HTTP annotations for gRPC Gateway

### gRPC Server
- Runs on port 9090
- Provides native gRPC API
- Can be tested with grpcurl

### gRPC Gateway
- Runs on port 8080 (configurable)
- Provides REST API from gRPC service
- Automatic code generation

### Code Generation
```bash
# Generate protobuf code
make proto
# or
buf generate
```

### Testing gRPC
```bash
# Test gRPC endpoints
make test-grpc
# or
./test_grpc.sh
```

## Monitoring and Observability

The application includes comprehensive monitoring and observability features:

### Prometheus Metrics
- HTTP request metrics (count, duration, status codes)
- Database connection and query metrics
- Kafka event publishing metrics
- System metrics (goroutines, memory usage)
- Business metrics (user count, events)

### Health Checks
- Database connectivity
- System health
- Kafka connectivity
- Custom health checks

### Grafana Dashboards
- Pre-configured dashboards for application metrics
- Real-time monitoring
- Alerting capabilities

### Monitoring Setup
```bash
# Start monitoring services
make monitoring-up

# Start HTTP server with monitoring
make run-http-new

# Test monitoring setup
./test_monitoring.sh
```

### Monitoring URLs
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics

## Development

### Adding New Features

1. Define domain entities in `internal/domain/entities/`
2. Create domain events in `internal/domain/events/`
3. Create repository interfaces in `internal/domain/repositories/`
4. Implement use cases in `internal/application/usecases/`
5. Add HTTP handlers in `internal/infrastructure/http/handlers/`
6. Update routes in `internal/infrastructure/http/server/server.go`

### Testing

```bash
go test ./...
```

### Building

```bash
make build
# or
go build -o bin/app main.go
```

### Running

The application now supports two modes using Cobra commands:

#### HTTP Server (gRPC Gateway)
```bash
make run
# or
make run-http
# or
./bin/app http
```

#### gRPC Server
```bash
make run-grpc
# or
./bin/app grpc
```

#### Available Commands
```bash
# Show all available commands
./bin/app --help

# Show HTTP command help
./bin/app http --help

# Show gRPC command help
./bin/app grpc --help
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License. 