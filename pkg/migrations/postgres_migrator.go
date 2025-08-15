package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// PostgresMigrator implements MigrationInterface for PostgreSQL
type PostgresMigrator struct {
	migrate *migrate.Migrate
}

// NewPostgresMigrator creates a new PostgreSQL migrator
func NewPostgresMigrator(db *sql.DB, migrationsPath string) (*PostgresMigrator, error) {
	// Create postgres driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create file source
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &PostgresMigrator{migrate: m}, nil
}

// Initialize creates the migrations table if it doesn't exist
func (p *PostgresMigrator) Initialize(ctx context.Context) error {
	// golang-migrate automatically creates the schema_migrations table
	// No additional initialization needed
	return nil
}

// Up runs all pending migrations
func (p *PostgresMigrator) Up(ctx context.Context) error {
	if err := p.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	return nil
}

// Down rolls back all migrations
func (p *PostgresMigrator) Down(ctx context.Context) error {
	if err := p.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations down: %w", err)
	}
	return nil
}

// Force sets the migration version (useful for fixing dirty state)
func (p *PostgresMigrator) Force(ctx context.Context, version int) error {
	if err := p.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}
	return nil
}

// Version returns the current migration version
func (p *PostgresMigrator) Version(ctx context.Context) (uint, bool, error) {
	version, dirty, err := p.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, err
}

// Steps runs n migrations (positive for up, negative for down)
func (p *PostgresMigrator) Steps(ctx context.Context, n int) error {
	if err := p.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration steps: %w", err)
	}
	return nil
}

// Close closes the migrator
func (p *PostgresMigrator) Close() error {
	if sourceErr, databaseErr := p.migrate.Close(); sourceErr != nil || databaseErr != nil {
		if sourceErr != nil {
			return fmt.Errorf("failed to close migrator source: %w", sourceErr)
		}
		return fmt.Errorf("failed to close migrator database: %w", databaseErr)
	}
	return nil
}
