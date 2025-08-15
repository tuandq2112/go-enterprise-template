package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/internal/infrastructure/messagebroker"
)

// WorkerPoolEventPublisher implements EventPublisher using worker pool for concurrent publishing
type WorkerPoolEventPublisher struct {
	broker     messagebroker.MessageBroker
	config     *config.Config
	workerPool []*PublisherWorker
	jobQueue   chan *PublishJob
	stopChan   chan struct{}
	wg         sync.WaitGroup
	metrics    *PublisherMetrics
}

// PublisherWorker represents a worker in the publisher pool
type PublisherWorker struct {
	id       int
	jobQueue <-chan *PublishJob
	broker   messagebroker.MessageBroker
	config   *config.Config
	stopChan <-chan struct{}
	wg       *sync.WaitGroup
	metrics  *PublisherMetrics
}

// PublishJob represents a job to publish an event
type PublishJob struct {
	Event      *events.Event
	Topic      string
	RetryCount int
	MaxRetries int
}

// PublisherMetrics holds metrics for the publisher
type PublisherMetrics struct {
	mu              sync.RWMutex
	PublishedEvents int64
	FailedEvents    int64
	RetryEvents     int64
	WorkerStats     map[int]*WorkerStats
}

// WorkerStats holds statistics for individual workers
type WorkerStats struct {
	JobsProcessed int64
	JobsFailed    int64
	LastJobTime   time.Time
}

// NewWorkerPoolEventPublisher creates a new worker pool event publisher
func NewWorkerPoolEventPublisher(broker messagebroker.MessageBroker, config *config.Config) *WorkerPoolEventPublisher {
	publisher := &WorkerPoolEventPublisher{
		broker:   broker,
		config:   config,
		jobQueue: make(chan *PublishJob, config.MessageBroker.WorkerBufferSize),
		stopChan: make(chan struct{}),
		metrics:  &PublisherMetrics{WorkerStats: make(map[int]*WorkerStats)},
	}

	// Create worker pool
	publisher.createWorkerPool()

	return publisher
}

// createWorkerPool creates the worker pool
func (p *WorkerPoolEventPublisher) createWorkerPool() {
	numWorkers := p.config.MessageBroker.PublisherWorkers
	if numWorkers <= 0 {
		numWorkers = 5 // Default to 5 workers
	}

	p.workerPool = make([]*PublisherWorker, numWorkers)

	for i := 0; i < numWorkers; i++ {
		worker := &PublisherWorker{
			id:       i + 1,
			jobQueue: p.jobQueue,
			broker:   p.broker,
			config:   p.config,
			stopChan: p.stopChan,
			wg:       &p.wg,
			metrics:  p.metrics,
		}

		p.workerPool[i] = worker
		p.wg.Add(1)

		// Initialize worker stats
		p.metrics.mu.Lock()
		p.metrics.WorkerStats[worker.id] = &WorkerStats{}
		p.metrics.mu.Unlock()

		// Start worker
		go worker.start()
	}

	log.Printf("Created publisher worker pool with %d workers", numWorkers)
}

// start starts the worker
func (w *PublisherWorker) start() {
	defer w.wg.Done()

	log.Printf("Publisher worker %d started", w.id)

	for {
		select {
		case <-w.stopChan:
			log.Printf("Publisher worker %d stopping", w.id)
			return
		case job := <-w.jobQueue:
			if job == nil {
				continue
			}

			w.processJob(job)
		}
	}
}

// processJob processes a publish job with retry logic
func (w *PublisherWorker) processJob(job *PublishJob) {
	startTime := time.Now()

	// Update worker stats
	w.metrics.mu.Lock()
	stats := w.metrics.WorkerStats[w.id]
	stats.JobsProcessed++
	stats.LastJobTime = startTime
	w.metrics.mu.Unlock()

	// Serialize event to JSON
	eventData, err := json.Marshal(job.Event)
	if err != nil {
		w.handleJobError(job, fmt.Errorf("failed to marshal event: %w", err))
		return
	}

	// Publish with retry logic
	var lastErr error
	for attempt := job.RetryCount; attempt <= job.MaxRetries; attempt++ {
		if err := w.broker.Publish(job.Topic, eventData); err == nil {
			// Success
			w.metrics.mu.Lock()
			w.metrics.PublishedEvents++
			w.metrics.mu.Unlock()

			log.Printf("Worker %d: Successfully published event %s to topic %s (attempt %d)",
				w.id, job.Event.Type, job.Topic, attempt)
			return
		} else {
			lastErr = err
			if attempt < job.MaxRetries {
				// Exponential backoff
				backoff := time.Duration(attempt) * time.Second
				log.Printf("Worker %d: Failed to publish event %s (attempt %d), retrying in %v: %v",
					w.id, job.Event.Type, attempt, backoff, err)
				time.Sleep(backoff)
			}
		}
	}

	// All attempts failed
	w.handleJobError(job, lastErr)
}

// handleJobError handles job processing errors
func (w *PublisherWorker) handleJobError(job *PublishJob, err error) {
	w.metrics.mu.Lock()
	w.metrics.FailedEvents++
	w.metrics.WorkerStats[w.id].JobsFailed++
	w.metrics.mu.Unlock()

	log.Printf("Worker %d: Failed to publish event %s to topic %s after %d attempts: %v",
		w.id, job.Event.Type, job.Topic, job.MaxRetries, err)
}

// PublishEvent publishes an event using the worker pool
func (p *WorkerPoolEventPublisher) PublishEvent(ctx context.Context, event *events.Event) error {
	// Get topic from config mapping
	topic := p.getTopicForEvent(event.Type)

	// Create job
	job := &PublishJob{
		Event:      event,
		Topic:      topic,
		RetryCount: 1,
		MaxRetries: 3,
	}

	// Send job to worker pool
	select {
	case p.jobQueue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full, try to publish directly
		return p.publishDirectly(ctx, event, topic)
	}
}

// publishDirectly publishes an event directly when worker pool is full
func (p *WorkerPoolEventPublisher) publishDirectly(ctx context.Context, event *events.Event, topic string) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.broker.Publish(topic, eventData)
}

// PublishEvents publishes multiple events using the worker pool
func (p *WorkerPoolEventPublisher) PublishEvents(ctx context.Context, events []*events.Event) error {
	for _, event := range events {
		if err := p.PublishEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// getTopicForEvent returns the appropriate topic for an event type
func (p *WorkerPoolEventPublisher) getTopicForEvent(eventType string) string {
	// Check if there's a mapping in config
	if mappedTopic, exists := p.config.MessageBroker.Topics[eventType]; exists {
		return mappedTopic
	}

	// Fallback to event type as topic name
	return eventType
}

// GetMetrics returns publisher metrics
func (p *WorkerPoolEventPublisher) GetMetrics() *PublisherMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &PublisherMetrics{
		PublishedEvents: p.metrics.PublishedEvents,
		FailedEvents:    p.metrics.FailedEvents,
		RetryEvents:     p.metrics.RetryEvents,
		WorkerStats:     make(map[int]*WorkerStats),
	}

	for id, stats := range p.metrics.WorkerStats {
		metrics.WorkerStats[id] = &WorkerStats{
			JobsProcessed: stats.JobsProcessed,
			JobsFailed:    stats.JobsFailed,
			LastJobTime:   stats.LastJobTime,
		}
	}

	return metrics
}

// Stop stops the worker pool
func (p *WorkerPoolEventPublisher) Stop() {
	log.Printf("Stopping publisher worker pool...")
	close(p.stopChan)
	p.wg.Wait()
	log.Printf("Publisher worker pool stopped")
}
