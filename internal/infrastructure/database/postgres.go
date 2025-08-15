package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-clean-ddd-es-template/internal/infrastructure/config"

	_ "github.com/lib/pq"
)

// PostgresDB represents PostgreSQL database connection
type PostgresDB struct {
	config *config.DatabaseConfig
	DB     *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	db := &PostgresDB{
		config: cfg,
	}

	if err := db.Connect(); err != nil {
		return nil, err
	}

	return db, nil
}

// NewPostgresConnection creates a new PostgreSQL connection and returns *sql.DB
func NewPostgresConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	configureConnectionPool(db, &cfg)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Printf("Connected to PostgreSQL database: %s", cfg.DBName)
	return db, nil
}

// Connect establishes connection to PostgreSQL database
func (p *PostgresDB) Connect() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		p.config.Host, p.config.Port, p.config.User, p.config.Password, p.config.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	configureConnectionPool(db, p.config)

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.DB = db
	log.Printf("Connected to PostgreSQL database: %s", p.config.DBName)
	return nil
}

// configureConnectionPool configures the database connection pool settings
func configureConnectionPool(db *sql.DB, cfg *config.DatabaseConfig) {
	// Set maximum number of open connections
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
		log.Printf("Set MaxOpenConns to %d", cfg.MaxOpenConns)
	}

	// Set maximum number of idle connections
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
		log.Printf("Set MaxIdleConns to %d", cfg.MaxIdleConns)
	}

	// Set maximum lifetime of connections
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		log.Printf("Set ConnMaxLifetime to %v", cfg.ConnMaxLifetime)
	}

	// Set maximum idle time of connections
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
		log.Printf("Set ConnMaxIdleTime to %v", cfg.ConnMaxIdleTime)
	}
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

// GetDB returns the underlying database connection
func (p *PostgresDB) GetDB() interface{} {
	return p.DB
}
