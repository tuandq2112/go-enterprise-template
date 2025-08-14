package database

import (
	"fmt"
	"log"

	"go-clean-ddd-es-template/internal/infrastructure/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresDB represents PostgreSQL database connection
type PostgresDB struct {
	config *config.DatabaseConfig
	DB     *gorm.DB
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

// Connect establishes connection to PostgreSQL database
func (p *PostgresDB) Connect() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		p.config.Host, p.config.Port, p.config.User, p.config.Password, p.config.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	p.DB = db
	log.Printf("Connected to PostgreSQL database: %s", p.config.DBName)
	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.DB != nil {
		sqlDB, err := p.DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB returns the underlying database connection
func (p *PostgresDB) GetDB() interface{} {
	return p.DB
}
