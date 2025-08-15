package workerpool

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job represents a generic job to be processed
type Job interface {
	Execute(ctx context.Context) error
	GetID() string
	GetRetryCount() int
	GetMaxRetries() int
	IncrementRetryCount()
}

// Worker represents a worker in the pool
type Worker struct {
	id       int
	jobQueue <-chan Job
	stopChan <-chan struct{}
	wg       *sync.WaitGroup
	metrics  *Metrics
	handler  JobHandler
}

// JobHandler defines how jobs should be processed
type JobHandler interface {
	ProcessJob(ctx context.Context, job Job) error
	HandleJobError(job Job, err error)
}

// Metrics holds worker pool metrics
type Metrics struct {
	mu            sync.RWMutex
	ProcessedJobs int64
	FailedJobs    int64
	RetryJobs     int64
	WorkerStats   map[int]*WorkerStats
}

// WorkerStats holds statistics for individual workers
type WorkerStats struct {
	JobsProcessed int64
	JobsFailed    int64
	LastJobTime   time.Time
}

// WorkerPool represents a pool of workers
type WorkerPool struct {
	workers    []*Worker
	jobQueue   chan Job
	stopChan   chan struct{}
	wg         sync.WaitGroup
	metrics    *Metrics
	handler    JobHandler
	numWorkers int
	bufferSize int
}

// Config holds worker pool configuration
type Config struct {
	NumWorkers int           // Number of workers in the pool
	BufferSize int           // Buffer size for job queue
	Handler    JobHandler    // Job handler implementation
	RetryDelay time.Duration // Delay between retries
	MaxRetries int           // Maximum number of retries per job
}

// DefaultConfig returns default worker pool configuration
func DefaultConfig() Config {
	return Config{
		NumWorkers: 10,
		BufferSize: 1000,
		RetryDelay: time.Second,
		MaxRetries: 3,
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config Config) *WorkerPool {
	if config.NumWorkers <= 0 {
		config.NumWorkers = 10
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}

	pool := &WorkerPool{
		jobQueue:   make(chan Job, config.BufferSize),
		stopChan:   make(chan struct{}),
		metrics:    &Metrics{WorkerStats: make(map[int]*WorkerStats)},
		handler:    config.Handler,
		numWorkers: config.NumWorkers,
		bufferSize: config.BufferSize,
	}

	pool.createWorkers()
	return pool
}

// createWorkers creates the worker pool
func (wp *WorkerPool) createWorkers() {
	wp.workers = make([]*Worker, wp.numWorkers)

	for i := 0; i < wp.numWorkers; i++ {
		worker := &Worker{
			id:       i + 1,
			jobQueue: wp.jobQueue,
			stopChan: wp.stopChan,
			wg:       &wp.wg,
			metrics:  wp.metrics,
			handler:  wp.handler,
		}

		wp.workers[i] = worker
		wp.wg.Add(1)

		// Initialize worker stats
		wp.metrics.mu.Lock()
		wp.metrics.WorkerStats[worker.id] = &WorkerStats{}
		wp.metrics.mu.Unlock()

		// Start worker
		go worker.start()
	}

	log.Printf("Created worker pool with %d workers", wp.numWorkers)
}

// start starts the worker
func (w *Worker) start() {
	defer w.wg.Done()

	log.Printf("Worker %d started", w.id)

	for {
		select {
		case <-w.stopChan:
			log.Printf("Worker %d stopping", w.id)
			return
		case job := <-w.jobQueue:
			if job == nil {
				continue
			}

			w.processJob(job)
		}
	}
}

// processJob processes a job with retry logic
func (w *Worker) processJob(job Job) {
	startTime := time.Now()

	// Update worker stats
	w.metrics.mu.Lock()
	stats := w.metrics.WorkerStats[w.id]
	stats.JobsProcessed++
	stats.LastJobTime = startTime
	w.metrics.mu.Unlock()

	// Process job with retry logic
	ctx := context.Background()
	var lastErr error

	for attempt := job.GetRetryCount(); attempt <= job.GetMaxRetries(); attempt++ {
		if err := w.handler.ProcessJob(ctx, job); err == nil {
			// Success
			w.metrics.mu.Lock()
			w.metrics.ProcessedJobs++
			w.metrics.mu.Unlock()

			log.Printf("Worker %d: Successfully processed job %s (attempt %d)",
				w.id, job.GetID(), attempt)
			return
		} else {
			lastErr = err
			if attempt < job.GetMaxRetries() {
				// Increment retry count
				job.IncrementRetryCount()

				// Exponential backoff
				backoff := time.Duration(attempt) * time.Second
				log.Printf("Worker %d: Failed to process job %s (attempt %d), retrying in %v: %v",
					w.id, job.GetID(), attempt, backoff, err)
				time.Sleep(backoff)
			}
		}
	}

	// All attempts failed
	w.handleJobError(job, lastErr)
}

// handleJobError handles job processing errors
func (w *Worker) handleJobError(job Job, err error) {
	w.metrics.mu.Lock()
	w.metrics.FailedJobs++
	w.metrics.WorkerStats[w.id].JobsFailed++
	w.metrics.mu.Unlock()

	// Let the handler deal with the error
	w.handler.HandleJobError(job, err)
}

// SubmitJob submits a job to the worker pool
func (wp *WorkerPool) SubmitJob(ctx context.Context, job Job) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full, try to process directly
		return wp.processDirectly(ctx, job)
	}
}

// processDirectly processes a job directly when worker pool is full
func (wp *WorkerPool) processDirectly(ctx context.Context, job Job) error {
	return wp.handler.ProcessJob(ctx, job)
}

// GetMetrics returns worker pool metrics
func (wp *WorkerPool) GetMetrics() *Metrics {
	wp.metrics.mu.RLock()
	defer wp.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &Metrics{
		ProcessedJobs: wp.metrics.ProcessedJobs,
		FailedJobs:    wp.metrics.FailedJobs,
		RetryJobs:     wp.metrics.RetryJobs,
		WorkerStats:   make(map[int]*WorkerStats),
	}

	for id, stats := range wp.metrics.WorkerStats {
		metrics.WorkerStats[id] = &WorkerStats{
			JobsProcessed: stats.JobsProcessed,
			JobsFailed:    stats.JobsFailed,
			LastJobTime:   stats.LastJobTime,
		}
	}

	return metrics
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	log.Printf("Stopping worker pool...")
	close(wp.stopChan)
	wp.wg.Wait()
	log.Printf("Worker pool stopped")
}

// GetStats returns worker pool statistics
func (wp *WorkerPool) GetStats() map[string]interface{} {
	metrics := wp.GetMetrics()

	stats := map[string]interface{}{
		"num_workers":    wp.numWorkers,
		"buffer_size":    wp.bufferSize,
		"processed_jobs": metrics.ProcessedJobs,
		"failed_jobs":    metrics.FailedJobs,
		"retry_jobs":     metrics.RetryJobs,
		"worker_stats":   metrics.WorkerStats,
	}

	return stats
}
