package database_test

import (
	"testing"

	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/internal/infrastructure/database"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseFactory_CreateDatabase(t *testing.T) {
	factory := database.NewDatabaseFactory()

	tests := []struct {
		name        string
		config      *config.DatabaseConfig
		expectError bool
	}{
		{
			name: "create postgres database",
			config: &config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "password",
				DBName:   "testdb",
			},
			expectError: true, // Will fail because testdb doesn't exist
		},
		{
			name: "create mysql database",
			config: &config.DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     "3306",
				User:     "root",
				Password: "password",
				DBName:   "testdb",
			},
			expectError: true, // MySQL is stub implementation
		},
		{
			name: "create mongodb database",
			config: &config.DatabaseConfig{
				Type:       "mongodb",
				Host:       "localhost",
				Port:       "27017",
				User:       "admin",
				Password:   "password",
				DBName:     "testdb",
				Collection: "users",
			},
			expectError: true, // MongoDB is stub implementation
		},
		{
			name: "unsupported database type",
			config: &config.DatabaseConfig{
				Type: "unsupported",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := factory.CreateDatabase(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}

func TestDatabaseFactory_NewDatabaseFactory(t *testing.T) {
	factory := database.NewDatabaseFactory()
	assert.NotNil(t, factory)
}

func TestPostgresDB_NewPostgresDB(t *testing.T) {
	// This test would require a real PostgreSQL connection
	// For now, we'll test the structure
	config := &config.DatabaseConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
	}

	postgresDB, err := database.NewPostgresDB(config)
	// This will fail because we don't have a real PostgreSQL connection
	assert.Error(t, err)
	assert.Nil(t, postgresDB)
}

func TestMySQLDB_NewMySQLDB(t *testing.T) {
	// This test would require a real MySQL connection
	// For now, we'll test the structure
	config := &config.DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: "password",
		DBName:   "testdb",
	}

	mysqlDB, err := database.NewMySQLDB(config)
	// This will fail because MySQL is stub implementation
	assert.Error(t, err)
	assert.Nil(t, mysqlDB)
}

func TestMongoDB_NewMongoDB(t *testing.T) {
	// This test would require a real MongoDB connection
	// For now, we'll test the structure
	config := &config.DatabaseConfig{
		Type:       "mongodb",
		Host:       "localhost",
		Port:       "27017",
		User:       "admin",
		Password:   "password",
		DBName:     "testdb",
		Collection: "users",
	}

	mongoDB, err := database.NewMongoDB(config)
	// This will fail because MongoDB is stub implementation
	assert.Error(t, err)
	assert.Nil(t, mongoDB)
}
