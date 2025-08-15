package repositories

import (
	"fmt"

	"go-clean-ddd-es-template/internal/domain/repositories"
	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/internal/infrastructure/database"
	"go-clean-ddd-es-template/internal/infrastructure/messagebroker"

	"go.mongodb.org/mongo-driver/mongo"
)

// RepositoryFactory creates repositories based on configuration
type RepositoryFactory struct {
	writeDB database.Database
	readDB  database.Database
	eventDB database.Database
	config  *config.Config
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(writeDB database.Database, readDB database.Database, eventDB database.Database, config *config.Config) *RepositoryFactory {
	return &RepositoryFactory{
		writeDB: writeDB,
		readDB:  readDB,
		eventDB: eventDB,
		config:  config,
	}
}

// CreateUserWriteRepository creates user write repository based on config
func (f *RepositoryFactory) CreateUserWriteRepository() (repositories.UserWriteRepository, error) {
	switch f.config.WriteDatabase.Type {
	case "postgres":
		return NewPostgresUserWriteRepository(f.writeDB), nil
	default:
		return nil, fmt.Errorf("unsupported write database type: %s", f.config.WriteDatabase.Type)
	}
}

// CreateUserReadRepository creates user read repository based on config
func (f *RepositoryFactory) CreateUserReadRepository() (repositories.UserReadRepository, error) {
	switch f.config.ReadDatabase.Type {
	case "mongodb":
		client := f.readDB.GetDB().(*mongo.Client)
		return NewMongoUserReadRepository(client, f.config.ReadDatabase.DBName, f.config.ReadDatabase.Collection), nil
	case "postgres":
		return NewPostgresUserReadRepository(f.readDB), nil
	default:
		return nil, fmt.Errorf("unsupported read database type: %s", f.config.ReadDatabase.Type)
	}
}

// CreateEventStore creates event store based on config
func (f *RepositoryFactory) CreateEventStore() (repositories.EventStore, error) {
	switch f.config.EventDatabase.Type {
	case "postgres":
		return NewPostgresEventStore(f.eventDB.GetDB()), nil
	default:
		return nil, fmt.Errorf("unsupported event store database type: %s", f.config.EventDatabase.Type)
	}
}

// CreateEventPublisher creates event publisher based on config
func (f *RepositoryFactory) CreateEventPublisher(broker interface{}) (repositories.EventPublisher, error) {
	// Cast broker to MessageBroker interface
	messageBroker, ok := broker.(messagebroker.MessageBroker)
	if !ok {
		return nil, fmt.Errorf("invalid broker type for event publisher")
	}

	// Create worker pool event publisher
	return NewWorkerPoolEventPublisher(messageBroker, f.config), nil
}
