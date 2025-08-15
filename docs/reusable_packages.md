# Reusable Packages (pkg/)

This document describes the reusable packages that have been extracted from the infrastructure layer and moved to the `pkg/` directory for better reusability across different projects.

## Overview

The following packages have been identified as reusable components and moved to `pkg/`:

1. **Worker Pool Pattern** (`pkg/workerpool/`)
2. **Event Processing** (`pkg/eventprocessor/`)
3. **Message Broker Interface** (`pkg/messagebroker/`)
4. **Database Connection Pool** (`pkg/database/connection_pool.go`)

## 1. Worker Pool Pattern (`pkg/workerpool/`)

A generic worker pool implementation that can be used for any concurrent job processing.

### Features

- **Configurable Worker Count**: Set the number of workers in the pool
- **Job Queue with Buffer**: Configurable buffer size for job queue
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **Metrics Collection**: Track processed jobs, failed jobs, and worker statistics
- **Graceful Shutdown**: Proper cleanup when stopping the pool

### Usage Example

```go
// Create job handler
type MyJobHandler struct{}

func (h *MyJobHandler) ProcessJob(ctx context.Context, job workerpool.Job) error {
    // Process the job
    return nil
}

func (h *MyJobHandler) HandleJobError(job workerpool.Job, err error) {
    // Handle job processing errors
}

// Create worker pool
config := workerpool.Config{
    NumWorkers:  10,
    BufferSize:  1000,
    Handler:     &MyJobHandler{},
    MaxRetries:  3,
}

pool := workerpool.NewWorkerPool(config)

// Submit jobs
job := &MyJob{ID: "job-1"}
err := pool.SubmitJob(ctx, job)

// Get metrics
metrics := pool.GetMetrics()
fmt.Printf("Processed jobs: %d\n", metrics.ProcessedJobs)

// Stop pool
pool.Stop()
```

## 2. Event Processing (`pkg/eventprocessor/`)

A generic event processing system that can handle different types of events with multiple handlers.

### Features

- **Multiple Event Handlers**: Register different handlers for different event types
- **Retry Logic**: Built-in retry mechanism for failed event processing
- **Metrics Collection**: Track processed events, failed events, and handler statistics
- **Raw Event Processing**: Process raw events from message brokers
- **Thread-Safe**: Safe for concurrent access

### Usage Example

```go
// Create event handler
type UserEventHandler struct{}

func (h *UserEventHandler) HandleEvent(ctx context.Context, event eventprocessor.Event) error {
    // Handle user event
    return nil
}

func (h *UserEventHandler) GetEventType() string {
    return "user.created"
}

// Create event processor
config := eventprocessor.DefaultConfig()
logger := &MyLogger{}
processor := eventprocessor.NewEventProcessor(config, logger)

// Register handlers
processor.RegisterHandler(&UserEventHandler{})

// Process events
event := &eventprocessor.GenericEvent{
    Type: "user.created",
    Data: map[string]interface{}{"user_id": "123"},
}

err := processor.ProcessEvent(ctx, event)

// Process raw events
rawEvent := []byte(`{"user_id": "123", "email": "test@example.com"}`)
err = processor.ProcessRawEvent(ctx, rawEvent, "user.created")
```

## 3. Message Broker Interface (`pkg/messagebroker/`)

A generic interface for message brokers that can be implemented for different message broker systems.

### Features

- **Generic Interface**: Works with any message broker implementation
- **Message Builder**: Fluent API for building messages
- **Consumer Groups**: Support for consumer group management
- **Topic Management**: Create, delete, and check topic existence
- **Health Monitoring**: Health checks and statistics collection

### Usage Example

```go
// Create message
message := messagebroker.NewMessageBuilder().
    WithTopic("user-events").
    WithKey([]byte("user-123")).
    WithValue([]byte(`{"user_id": "123"}`)).
    WithHeader("event_type", []byte("user.created")).
    Build()

// Publish message
err := broker.Publish(ctx, "user-events", message)

// Subscribe to topic
handler := func(ctx context.Context, msg *messagebroker.Message) error {
    // Handle message
    return nil
}

err = broker.Subscribe(ctx, "user-events", handler)

// Create consumer group
consumerGroup := messagebroker.NewConsumerGroup("user-service", []string{"user-events"}, handler)
err = broker.JoinConsumerGroup(ctx, "user-service", []string{"user-events"})
```

## 4. Database Connection Pool (`pkg/database/connection_pool.go`)

A generic connection pool implementation that can be used with any database system.

### Features

- **Generic Interface**: Works with any database connection type
- **Connection Factory**: Pluggable connection creation
- **Health Checking**: Periodic health checks on connections
- **Statistics**: Detailed connection pool statistics
- **Configurable**: Configurable pool size, timeouts, and limits

### Usage Example

```go
// Create connection factory
type PostgresConnectionFactory struct{}

func (f *PostgresConnectionFactory) CreateConnection(ctx context.Context) (database.Connection, error) {
    // Create PostgreSQL connection
    return &PostgresConnection{}, nil
}

func (f *PostgresConnectionFactory) ValidateConnection(ctx context.Context, conn database.Connection) error {
    return conn.Ping(ctx)
}

// Create connection pool
config := database.DefaultPoolConfig()
config.MaxOpenConns = 25
config.MaxIdleConns = 5

factory := &PostgresConnectionFactory{}
pool := database.NewConnectionPool(factory, config)

// Get connection
conn, err := pool.GetConnection(ctx)
if err != nil {
    return err
}
defer pool.ReturnConnection(conn)

// Use connection
err = conn.Ping(ctx)

// Get statistics
stats := pool.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
```

## Benefits of Moving to pkg/

### 1. **Reusability**
- These packages can be used in other projects without modification
- No tight coupling to specific infrastructure implementations

### 2. **Testability**
- Easier to unit test in isolation
- Mock implementations can be created easily

### 3. **Maintainability**
- Clear separation of concerns
- Easier to maintain and update

### 4. **Flexibility**
- Can be used with different implementations
- Configurable for different use cases

### 5. **Performance**
- Optimized for specific use cases
- Can be tuned for different performance requirements

## Migration Guide

### From Infrastructure to pkg/

1. **Identify Reusable Components**: Look for patterns that could be used in other projects
2. **Extract Interfaces**: Create generic interfaces that are not tied to specific implementations
3. **Move to pkg/**: Move the reusable code to appropriate packages in `pkg/`
4. **Update Dependencies**: Update the infrastructure layer to use the new pkg/ components
5. **Add Documentation**: Document the new packages and their usage

### Example Migration

**Before (in infrastructure):**
```go
// internal/infrastructure/consumers/event_consumer.go
type EventConsumer struct {
    handlers map[string]EventHandler
    // ... specific to this implementation
}
```

**After (in pkg/):**
```go
// pkg/eventprocessor/event_processor.go
type EventProcessor struct {
    handlers map[string]EventHandler
    // ... generic implementation
}

// internal/infrastructure/consumers/event_consumer.go
type EventConsumer struct {
    processor *eventprocessor.EventProcessor
    // ... uses the generic processor
}
```

## Future Enhancements

1. **More Generic Patterns**: Extract more reusable patterns from infrastructure
2. **Plugin System**: Create plugin interfaces for extensibility
3. **Configuration Management**: Centralized configuration management
4. **Monitoring Integration**: Better integration with monitoring systems
5. **Performance Optimization**: Optimize for high-performance scenarios 