package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// MigrationInterface defines the interface for database migrations
type MigrationInterface interface {
	// Initialize creates the migrations table if it doesn't exist
	Initialize(ctx context.Context) error

	// Up runs all pending migrations
	Up(ctx context.Context) error

	// Down rolls back all migrations
	Down(ctx context.Context) error

	// Version returns the current migration version
	Version(ctx context.Context) (uint, bool, error)

	// Steps runs n migrations (positive for up, negative for down)
	Steps(ctx context.Context, n int) error

	// Force sets the migration version (useful for fixing dirty state)
	Force(ctx context.Context, version int) error

	// Close closes the migrator
	Close() error
}

// MigrationManager manages migrations for different database types
type MigrationManager struct {
	WriteDBMigrator MigrationInterface
	EventDBMigrator MigrationInterface
	ReadDBMigrator  MigrationInterface
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(
	writeDB *sql.DB,
	eventDB *sql.DB,
	writeMigrationsPath string,
	eventMigrationsPath string,
) (*MigrationManager, error) {
	// Create PostgreSQL migrators for write and event databases
	writeMigrator, err := NewPostgresMigrator(writeDB, writeMigrationsPath)
	if err != nil {
		return nil, err
	}

	eventMigrator, err := NewPostgresMigrator(eventDB, eventMigrationsPath)
	if err != nil {
		return nil, err
	}

	return &MigrationManager{
		WriteDBMigrator: writeMigrator,
		EventDBMigrator: eventMigrator,
		ReadDBMigrator:  nil, // MongoDB doesn't need SQL migrations
	}, nil
}

// Initialize initializes all migration systems
func (m *MigrationManager) Initialize(ctx context.Context) error {
	// Initialize write database migrations
	if err := m.WriteDBMigrator.Initialize(ctx); err != nil {
		return err
	}

	// Initialize event database migrations
	if err := m.EventDBMigrator.Initialize(ctx); err != nil {
		return err
	}

	return nil
}

// RunWriteDBMigrations runs migrations for write database
func (m *MigrationManager) RunWriteDBMigrations(ctx context.Context) error {
	return m.WriteDBMigrator.Up(ctx)
}

// RunEventDBMigrations runs migrations for event database
func (m *MigrationManager) RunEventDBMigrations(ctx context.Context) error {
	return m.EventDBMigrator.Up(ctx)
}

// GetWriteDBVersion returns the current version of write database
func (m *MigrationManager) GetWriteDBVersion(ctx context.Context) (uint, bool, error) {
	return m.WriteDBMigrator.Version(ctx)
}

// GetEventDBVersion returns the current version of event database
func (m *MigrationManager) GetEventDBVersion(ctx context.Context) (uint, bool, error) {
	return m.EventDBMigrator.Version(ctx)
}

// Close closes all migrators
func (m *MigrationManager) Close() error {
	if err := m.WriteDBMigrator.Close(); err != nil {
		return err
	}
	if err := m.EventDBMigrator.Close(); err != nil {
		return err
	}
	return nil
}

// CreateMigrationFile creates a new migration file
func CreateMigrationFile(migrationsPath, name string) error {
	// Ensure migrations directory exists
	if err := os.MkdirAll(migrationsPath, 0o755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create up migration file
	upFile := filepath.Join(migrationsPath, fmt.Sprintf("%s.up.sql", name))
	if err := os.WriteFile(upFile, []byte("-- Migration: "+name+"\n-- Description: \n\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	downFile := filepath.Join(migrationsPath, fmt.Sprintf("%s.down.sql", name))
	if err := os.WriteFile(downFile, []byte("-- Migration: "+name+"\n-- Description: Rollback\n\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files: %s.up.sql, %s.down.sql\n", name, name)
	return nil
}
