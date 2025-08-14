package metrics

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec

	// Database metrics
	DBConnectionsActive *prometheus.GaugeVec
	DBQueryDuration     *prometheus.HistogramVec
	DBQueriesTotal      *prometheus.CounterVec

	// Kafka metrics
	KafkaEventsPublished *prometheus.CounterVec
	KafkaEventsFailed    *prometheus.CounterVec
	KafkaProducerErrors  *prometheus.CounterVec

	// Business metrics
	UsersTotal      *prometheus.GaugeVec
	EventsStored    *prometheus.CounterVec
	EventsPublished *prometheus.CounterVec

	// System metrics
	GoRoutines  *prometheus.GaugeVec
	MemoryAlloc *prometheus.GaugeVec
	MemoryHeap  *prometheus.GaugeVec
}

// NewMetrics creates and registers all metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
			[]string{"method", "endpoint"},
		),

		// Database metrics
		DBConnectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
			[]string{"database"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		// Kafka metrics
		KafkaEventsPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_events_published_total",
				Help: "Total number of events published to Kafka",
			},
			[]string{"topic", "event_type"},
		),
		KafkaEventsFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_events_failed_total",
				Help: "Total number of failed events",
			},
			[]string{"topic", "event_type", "error"},
		),
		KafkaProducerErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_producer_errors_total",
				Help: "Total number of Kafka producer errors",
			},
			[]string{"error"},
		),

		// Business metrics
		UsersTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "users_total",
				Help: "Total number of users in the system",
			},
			[]string{},
		),
		EventsStored: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "events_stored_total",
				Help: "Total number of events stored in event store",
			},
			[]string{"event_type", "aggregate_type"},
		),
		EventsPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "events_published_total",
				Help: "Total number of events published",
			},
			[]string{"event_type", "aggregate_type"},
		),

		// System metrics
		GoRoutines: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "go_goroutines",
				Help: "Number of goroutines",
			},
			[]string{},
		),
		MemoryAlloc: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "go_memory_alloc_bytes",
				Help: "Memory allocated in bytes",
			},
			[]string{},
		),
		MemoryHeap: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "go_memory_heap_bytes",
				Help: "Heap memory in bytes",
			},
			[]string{},
		),
	}

	return m
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, endpoint, status string, duration float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordHTTPRequestInFlight records in-flight HTTP requests
func (m *Metrics) RecordHTTPRequestInFlight(method, endpoint string, count float64) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Set(count)
}

// RecordDBQuery records database query metrics
func (m *Metrics) RecordDBQuery(operation, table, status string, duration float64) {
	m.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordDBConnections records database connection metrics
func (m *Metrics) RecordDBConnections(database string, count float64) {
	m.DBConnectionsActive.WithLabelValues(database).Set(count)
}

// RecordKafkaEventPublished records Kafka event published
func (m *Metrics) RecordKafkaEventPublished(topic, eventType string) {
	m.KafkaEventsPublished.WithLabelValues(topic, eventType).Inc()
}

// RecordKafkaEventFailed records Kafka event failure
func (m *Metrics) RecordKafkaEventFailed(topic, eventType, error string) {
	m.KafkaEventsFailed.WithLabelValues(topic, eventType, error).Inc()
}

// RecordKafkaProducerError records Kafka producer error
func (m *Metrics) RecordKafkaProducerError(error string) {
	m.KafkaProducerErrors.WithLabelValues(error).Inc()
}

// RecordUsersTotal records total users count
func (m *Metrics) RecordUsersTotal(count float64) {
	m.UsersTotal.WithLabelValues().Set(count)
}

// RecordEventStored records event stored in event store
func (m *Metrics) RecordEventStored(eventType, aggregateType string) {
	m.EventsStored.WithLabelValues(eventType, aggregateType).Inc()
}

// RecordEventPublished records event published
func (m *Metrics) RecordEventPublished(eventType, aggregateType string) {
	m.EventsPublished.WithLabelValues(eventType, aggregateType).Inc()
}

// UpdateSystemMetrics updates system metrics
func (m *Metrics) UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.GoRoutines.WithLabelValues().Set(float64(runtime.NumGoroutine()))
	m.MemoryAlloc.WithLabelValues().Set(float64(memStats.Alloc))
	m.MemoryHeap.WithLabelValues().Set(float64(memStats.HeapAlloc))
}
