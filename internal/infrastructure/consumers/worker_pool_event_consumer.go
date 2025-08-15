package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/pkg/resilience"

	"github.com/IBM/sarama"
)

// WorkerPoolEventConsumer handles event consumption with worker pool
type WorkerPoolEventConsumer struct {
	eventHandlers   map[string]EventHandler
	deadLetterQueue *resilience.DeadLetterQueue
	logger          Logger
	config          *config.Config
	consumer        sarama.Consumer
	workerPool      []*ConsumerWorker
	jobQueue        chan *ConsumeJob
	stopChan        chan struct{}
	wg              sync.WaitGroup
	metrics         *ConsumerMetrics
}

// ConsumerWorker represents a worker in the consumer pool
type ConsumerWorker struct {
	id       int
	jobQueue <-chan *ConsumeJob
	handlers map[string]EventHandler
	dlq      *resilience.DeadLetterQueue
	logger   Logger
	stopChan <-chan struct{}
	wg       *sync.WaitGroup
	metrics  *ConsumerMetrics
}

// ConsumeJob represents a job to consume an event
type ConsumeJob struct {
	Message    []byte
	Topic      string
	Partition  int32
	Offset     int64
	RetryCount int
	MaxRetries int
}

// ConsumerMetrics holds metrics for the consumer
type ConsumerMetrics struct {
	mu              sync.RWMutex
	ProcessedEvents int64
	FailedEvents    int64
	RetryEvents     int64
	WorkerStats     map[int]*ConsumerWorkerStats
}

// ConsumerWorkerStats holds statistics for individual consumer workers
type ConsumerWorkerStats struct {
	JobsProcessed int64
	JobsFailed    int64
	LastJobTime   time.Time
}

// NewWorkerPoolEventConsumer creates a new worker pool event consumer
func NewWorkerPoolEventConsumer(config *config.Config, consumer sarama.Consumer, logger Logger) *WorkerPoolEventConsumer {
	// Create dead letter queue with in-memory storage
	dlqConfig := resilience.DefaultDeadLetterQueueConfig()
	dlq := resilience.NewDeadLetterQueue(dlqConfig, nil, nil)

	eventConsumer := &WorkerPoolEventConsumer{
		eventHandlers:   make(map[string]EventHandler),
		deadLetterQueue: dlq,
		logger:          logger,
		config:          config,
		consumer:        consumer,
		jobQueue:        make(chan *ConsumeJob, config.MessageBroker.WorkerBufferSize),
		stopChan:        make(chan struct{}),
		metrics:         &ConsumerMetrics{WorkerStats: make(map[int]*ConsumerWorkerStats)},
	}

	// Create worker pool
	eventConsumer.createWorkerPool()

	return eventConsumer
}

// createWorkerPool creates the worker pool
func (ec *WorkerPoolEventConsumer) createWorkerPool() {
	numWorkers := ec.config.MessageBroker.ConsumerWorkers
	if numWorkers <= 0 {
		numWorkers = 10 // Default to 10 workers
	}

	ec.workerPool = make([]*ConsumerWorker, numWorkers)

	for i := 0; i < numWorkers; i++ {
		worker := &ConsumerWorker{
			id:       i + 1,
			jobQueue: ec.jobQueue,
			handlers: ec.eventHandlers,
			dlq:      ec.deadLetterQueue,
			logger:   ec.logger,
			stopChan: ec.stopChan,
			wg:       &ec.wg,
			metrics:  ec.metrics,
		}

		ec.workerPool[i] = worker
		ec.wg.Add(1)

		// Initialize worker stats
		ec.metrics.mu.Lock()
		ec.metrics.WorkerStats[worker.id] = &ConsumerWorkerStats{}
		ec.metrics.mu.Unlock()

		// Start worker
		go worker.start()
	}

	ec.logger.Info("Created consumer worker pool with %d workers", numWorkers)
}

// start starts the worker
func (w *ConsumerWorker) start() {
	defer w.wg.Done()

	w.logger.Info("Consumer worker %d started", w.id)

	for {
		select {
		case <-w.stopChan:
			w.logger.Info("Consumer worker %d stopping", w.id)
			return
		case job := <-w.jobQueue:
			if job == nil {
				continue
			}

			w.processJob(job)
		}
	}
}

// processJob processes a consume job with retry logic
func (w *ConsumerWorker) processJob(job *ConsumeJob) {
	startTime := time.Now()

	// Update worker stats
	w.metrics.mu.Lock()
	stats := w.metrics.WorkerStats[w.id]
	stats.JobsProcessed++
	stats.LastJobTime = startTime
	w.metrics.mu.Unlock()

	// Parse event from message
	var event events.Event
	if err := json.Unmarshal(job.Message, &event); err != nil {
		w.handleJobError(job, fmt.Errorf("failed to unmarshal event: %w", err))
		return
	}

	// Convert to UserEvent format for processing
	userEvent := &entities.UserEvent{
		UserID:    "", // Will be extracted from event data
		EventType: event.Type,
		EventData: make(map[string]interface{}),
		Timestamp: event.Timestamp,
		Version:   event.Version,
	}

	// Parse event data
	if len(event.Data) > 0 {
		if err := json.Unmarshal(event.Data, &userEvent.EventData); err != nil {
			w.handleJobError(job, fmt.Errorf("failed to unmarshal event data: %w", err))
			return
		}
	}

	// Extract user_id from event data
	if userID, ok := userEvent.EventData["user_id"].(string); ok {
		userEvent.UserID = userID
	}

	// Process the event with retry logic
	var lastErr error
	for attempt := job.RetryCount; attempt <= job.MaxRetries; attempt++ {
		if err := w.processEvent(userEvent); err == nil {
			// Success
			w.metrics.mu.Lock()
			w.metrics.ProcessedEvents++
			w.metrics.mu.Unlock()

			w.logger.Info("Worker %d: Successfully processed event %s from topic %s partition %d offset %d (attempt %d)",
				w.id, userEvent.EventType, job.Topic, job.Partition, job.Offset, attempt)
			return
		} else {
			lastErr = err
			if attempt < job.MaxRetries {
				// Exponential backoff
				backoff := time.Duration(attempt) * time.Second
				w.logger.Warn("Worker %d: Failed to process event %s (attempt %d), retrying in %v: %v",
					w.id, userEvent.EventType, attempt, backoff, err)
				time.Sleep(backoff)
			}
		}
	}

	// All attempts failed, add to dead letter queue
	w.handleJobError(job, lastErr)
}

// processEvent processes a single event
func (w *ConsumerWorker) processEvent(event *entities.UserEvent) error {
	// Find and execute handler
	handler, exists := w.handlers[event.EventType]
	if !exists {
		return fmt.Errorf("no handler registered for event type: %s", event.EventType)
	}

	// Execute handler
	return handler.HandleEvent(context.Background(), event)
}

// handleJobError handles job processing errors
func (w *ConsumerWorker) handleJobError(job *ConsumeJob, err error) {
	w.metrics.mu.Lock()
	w.metrics.FailedEvents++
	w.metrics.WorkerStats[w.id].JobsFailed++
	w.metrics.mu.Unlock()

	// Add to dead letter queue
	eventData := map[string]interface{}{
		"topic":     job.Topic,
		"partition": job.Partition,
		"offset":    job.Offset,
		"message":   string(job.Message),
	}

	metadata := map[string]string{
		"source": "worker_pool_consumer",
		"worker": fmt.Sprintf("%d", w.id),
		"error":  err.Error(),
	}

	if dlqErr := w.dlq.AddEvent(context.Background(), "failed_event", eventData, err, metadata); dlqErr != nil {
		w.logger.Error("Failed to add event to dead letter queue: %v", dlqErr)
	} else {
		w.logger.Warn("Event added to dead letter queue: %v, error: %v", eventData, err)
	}
}

// RegisterHandler registers an event handler for a specific event type
func (ec *WorkerPoolEventConsumer) RegisterHandler(eventType string, handler EventHandler) {
	ec.eventHandlers[eventType] = handler

	// Update handlers in all workers
	for _, worker := range ec.workerPool {
		worker.handlers = ec.eventHandlers
	}
}

// HandleMessage processes a message using the worker pool
func (ec *WorkerPoolEventConsumer) HandleMessage(ctx context.Context, message []byte) error {
	// Create job
	job := &ConsumeJob{
		Message:    message,
		Topic:      "unknown", // Will be set by the caller
		Partition:  0,
		Offset:     0,
		RetryCount: 1,
		MaxRetries: 3,
	}

	// Send job to worker pool
	select {
	case ec.jobQueue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full, try to process directly
		return ec.processDirectly(ctx, message)
	}
}

// processDirectly processes a message directly when worker pool is full
func (ec *WorkerPoolEventConsumer) processDirectly(ctx context.Context, message []byte) error {
	// Parse event from message
	var event events.Event
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Convert to UserEvent format for processing
	userEvent := &entities.UserEvent{
		UserID:    "",
		EventType: event.Type,
		EventData: make(map[string]interface{}),
		Timestamp: event.Timestamp,
		Version:   event.Version,
	}

	// Parse event data
	if len(event.Data) > 0 {
		if err := json.Unmarshal(event.Data, &userEvent.EventData); err != nil {
			return fmt.Errorf("failed to unmarshal event data: %w", err)
		}
	}

	// Extract user_id from event data
	if userID, ok := userEvent.EventData["user_id"].(string); ok {
		userEvent.UserID = userID
	}

	// Process the event
	return ec.processEvent(ctx, userEvent)
}

// processEvent processes a single event
func (ec *WorkerPoolEventConsumer) processEvent(ctx context.Context, event *entities.UserEvent) error {
	// Find and execute handler
	handler, exists := ec.eventHandlers[event.EventType]
	if !exists {
		return fmt.Errorf("no handler registered for event type: %s", event.EventType)
	}

	// Execute handler with retry logic
	return ec.executeWithRetry(ctx, func() error {
		return handler.HandleEvent(ctx, event)
	})
}

// executeWithRetry executes a function with retry logic
func (ec *WorkerPoolEventConsumer) executeWithRetry(ctx context.Context, fn func() error) error {
	maxAttempts := 3
	delay := time.Second

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxAttempts {
				ec.logger.Warn("Attempt %d failed, retrying in %v: %v", attempt, delay, err)
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// GetMetrics returns consumer metrics
func (ec *WorkerPoolEventConsumer) GetMetrics() *ConsumerMetrics {
	ec.metrics.mu.RLock()
	defer ec.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &ConsumerMetrics{
		ProcessedEvents: ec.metrics.ProcessedEvents,
		FailedEvents:    ec.metrics.FailedEvents,
		RetryEvents:     ec.metrics.RetryEvents,
		WorkerStats:     make(map[int]*ConsumerWorkerStats),
	}

	for id, stats := range ec.metrics.WorkerStats {
		metrics.WorkerStats[id] = &ConsumerWorkerStats{
			JobsProcessed: stats.JobsProcessed,
			JobsFailed:    stats.JobsFailed,
			LastJobTime:   stats.LastJobTime,
		}
	}

	return metrics
}

// GetDLQStats returns dead letter queue statistics
func (ec *WorkerPoolEventConsumer) GetDLQStats(ctx context.Context) (resilience.DLQStats, error) {
	return ec.deadLetterQueue.GetStats(ctx)
}

// RetryFailedEvent retries a failed event from dead letter queue
func (ec *WorkerPoolEventConsumer) RetryFailedEvent(ctx context.Context, eventID string) error {
	return ec.deadLetterQueue.RetryEvent(ctx, eventID)
}

// ListFailedEvents lists failed events from dead letter queue
func (ec *WorkerPoolEventConsumer) ListFailedEvents(ctx context.Context, limit, offset int) ([]*resilience.FailedEvent, error) {
	return ec.deadLetterQueue.ListEvents(ctx, limit, offset)
}

// GetFailedEvent gets a specific failed event from dead letter queue
func (ec *WorkerPoolEventConsumer) GetFailedEvent(ctx context.Context, eventID string) (*resilience.FailedEvent, error) {
	return ec.deadLetterQueue.GetEvent(ctx, eventID)
}

// DeleteFailedEvent deletes a failed event from dead letter queue
func (ec *WorkerPoolEventConsumer) DeleteFailedEvent(ctx context.Context, eventID string) error {
	return ec.deadLetterQueue.DeleteEvent(ctx, eventID)
}

// Stop stops the worker pool
func (ec *WorkerPoolEventConsumer) Stop() {
	ec.logger.Info("Stopping consumer worker pool...")
	close(ec.stopChan)
	ec.wg.Wait()
	ec.logger.Info("Consumer worker pool stopped")
}
