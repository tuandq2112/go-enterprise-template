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

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.DB = db
	log.Printf("Connected to PostgreSQL database: %s", p.config.DBName)
	return nil
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
