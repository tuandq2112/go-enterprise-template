package database

import (
	"context"
	"fmt"
	"time"

	"go-clean-ddd-es-template/internal/infrastructure/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database interface for different database types
type Database interface {
	Connect() error
	Close() error
	GetDB() interface{} // Returns the underlying database connection
}

// DatabaseFactory creates database instances based on configuration
type DatabaseFactory struct{}

// NewDatabaseFactory creates a new database factory
func NewDatabaseFactory() *DatabaseFactory {
	return &DatabaseFactory{}
}

// CreateDatabase creates a database instance based on configuration
func (f *DatabaseFactory) CreateDatabase(cfg *config.DatabaseConfig) (Database, error) {
	switch cfg.Type {
	case "postgres":
		return NewPostgresDB(cfg)
	case "mysql":
		return NewMySQLDB(cfg)
	case "mongodb":
		return NewMongoDB(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

// CreateMongoDB creates a MongoDB instance based on configuration
func (f *DatabaseFactory) CreateMongoDB(cfg *config.DatabaseConfig) (*MongoDB, error) {
	return NewMongoDB(cfg)
}

// MySQLDB stub implementation
type MySQLDB struct {
	config *config.DatabaseConfig
	DB     interface{}
}

func NewMySQLDB(cfg *config.DatabaseConfig) (*MySQLDB, error) {
	db := &MySQLDB{
		config: cfg,
	}

	if err := db.Connect(); err != nil {
		return nil, err
	}

	return db, nil
}

func (m *MySQLDB) Connect() error {
	// Stub implementation - would connect to MySQL in real implementation
	return fmt.Errorf("MySQL implementation not available - use PostgreSQL instead")
}

func (m *MySQLDB) Close() error {
	return nil
}

func (m *MySQLDB) GetDB() interface{} {
	return m.DB
}

// MongoDB implementation
type MongoDB struct {
	config *config.DatabaseConfig
	client *mongo.Client
}

func NewMongoDB(cfg *config.DatabaseConfig) (*MongoDB, error) {
	db := &MongoDB{
		config: cfg,
	}

	if err := db.Connect(); err != nil {
		return nil, err
	}

	return db, nil
}

func (m *MongoDB) Connect() error {
	// Build MongoDB URI if not provided
	uri := m.config.URI
	if uri == "" {
		uri = fmt.Sprintf("mongodb://%s:%s", m.config.Host, m.config.Port)
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	return nil
}

func (m *MongoDB) Close() error {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.client.Disconnect(ctx)
	}
	return nil
}

func (m *MongoDB) GetDB() interface{} {
	return m.client
}
