package repositories

import (
	"context"
	"fmt"
	"time"

	domainEvent "go-clean-ddd-es-template/internal/domain/events"
	"go-clean-ddd-es-template/internal/infrastructure/database"
)

// PostgresEventStore implements EventStore using PostgreSQL
type PostgresEventStore struct {
	db database.Database
}

// NewPostgresEventStore creates a new PostgreSQL event store
func NewPostgresEventStore(db interface{}) *PostgresEventStore {
	return &PostgresEventStore{
		db: &databaseWrapper{db: db},
	}
}

// databaseWrapper wraps the database connection to implement Database interface
type databaseWrapper struct {
	db interface{}
}

func (d *databaseWrapper) Connect() error {
	return nil // Already connected
}

func (d *databaseWrapper) Close() error {
	return nil // Will be handled by the actual database
}

func (d *databaseWrapper) GetDB() interface{} {
	return d.db
}

// SaveEvent saves an event to the event store
func (s *PostgresEventStore) SaveEvent(ctx context.Context, aggregateID string, event *domainEvent.Event) error {
	// Get underlying database connection
	dbConn := s.db.GetDB()
	if dbConn == nil {
		return fmt.Errorf("database connection not available")
	}

	// For now, we'll return an error indicating the implementation is not available
	// In a real implementation, you'd have type assertions to handle different database types
	return fmt.Errorf("event store implementation not available - use PostgreSQL")
}

// GetEvents retrieves all events for an aggregate
func (s *PostgresEventStore) GetEvents(ctx context.Context, aggregateID string) ([]*domainEvent.Event, error) {
	// Get underlying database connection
	dbConn := s.db.GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// For now, we'll return an error indicating the implementation is not available
	// In a real implementation, you'd have type assertions to handle different database types
	return nil, fmt.Errorf("event store implementation not available - use PostgreSQL")
}

// GetEventsByType retrieves events by type
func (s *PostgresEventStore) GetEventsByType(ctx context.Context, eventType string) ([]*domainEvent.Event, error) {
	// Get underlying database connection
	dbConn := s.db.GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// For now, we'll return an error indicating the implementation is not available
	// In a real implementation, you'd have type assertions to handle different database types
	return nil, fmt.Errorf("event store implementation not available - use PostgreSQL")
}

// GetEventsSince retrieves events since a specific timestamp
func (s *PostgresEventStore) GetEventsSince(ctx context.Context, since time.Time) ([]*domainEvent.Event, error) {
	// Get underlying database connection
	dbConn := s.db.GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// For now, we'll return an error indicating the implementation is not available
	// In a real implementation, you'd have type assertions to handle different database types
	return nil, fmt.Errorf("event store implementation not available - use PostgreSQL")
}

// GetLastEventVersion gets the last event version for an aggregate
func (s *PostgresEventStore) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	// Get underlying database connection
	dbConn := s.db.GetDB()
	if dbConn == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	// For now, we'll return an error indicating the implementation is not available
	// In a real implementation, you'd have type assertions to handle different database types
	return 0, fmt.Errorf("event store implementation not available - use PostgreSQL")
}

// Close closes the database connection
func (s *PostgresEventStore) Close() error {
	return s.db.Close()
}
