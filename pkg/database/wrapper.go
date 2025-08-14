package database

import (
	"database/sql"
	"time"

	"go-clean-ddd-es-template/pkg/metrics"
)

// DBWrapper wraps database operations with metrics
type DBWrapper struct {
	db      *sql.DB
	metrics *metrics.Metrics
}

// NewDBWrapper creates a new database wrapper
func NewDBWrapper(db *sql.DB, m *metrics.Metrics) *DBWrapper {
	return &DBWrapper{
		db:      db,
		metrics: m,
	}
}

// Query wraps sql.DB.Query with metrics
func (w *DBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.db.Query(query, args...)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("query", "users", status, duration)
	return rows, err
}

// QueryRow wraps sql.DB.QueryRow with metrics
func (w *DBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := w.db.QueryRow(query, args...)
	duration := time.Since(start).Seconds()

	// Note: QueryRow doesn't return error immediately, so we assume success
	w.metrics.RecordDBQuery("query_row", "users", "success", duration)
	return row
}

// Exec wraps sql.DB.Exec with metrics
func (w *DBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := w.db.Exec(query, args...)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("exec", "users", status, duration)
	return result, err
}

// Begin wraps sql.DB.Begin with metrics
func (w *DBWrapper) Begin() (*sql.Tx, error) {
	start := time.Now()
	tx, err := w.db.Begin()
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("begin", "transaction", status, duration)
	return tx, err
}

// Close wraps sql.DB.Close
func (w *DBWrapper) Close() error {
	return w.db.Close()
}

// Ping wraps sql.DB.Ping with metrics
func (w *DBWrapper) Ping() error {
	start := time.Now()
	err := w.db.Ping()
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("ping", "database", status, duration)
	return err
}

// Stats returns database stats
func (w *DBWrapper) Stats() sql.DBStats {
	stats := w.db.Stats()
	w.metrics.RecordDBConnections("postgres", float64(stats.OpenConnections))
	return stats
}
