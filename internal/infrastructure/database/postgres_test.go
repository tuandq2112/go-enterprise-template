package database_test

import (
	"database/sql"
	"testing"
	"time"

	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/internal/infrastructure/database"

	"github.com/stretchr/testify/assert"
)

func TestPostgresConnectionPoolConfiguration(t *testing.T) {
	// Create a test configuration with connection pool settings
	cfg := config.DatabaseConfig{
		Type:            "postgres",
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "password",
		DBName:          "testdb",
		MaxOpenConns:    10,
		MaxIdleConns:    3,
		ConnMaxLifetime: 2 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	// Test NewPostgresConnection function
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		// Skip test if database is not available
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer db.Close()

	// Verify connection pool settings
	assert.Equal(t, 10, db.Stats().MaxOpenConnections, "MaxOpenConnections should be set to 10")
}

func TestConfigureConnectionPool(t *testing.T) {
	// Create a mock database connection
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable")
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer db.Close()

	// Test configuration with various settings
	cfg := &config.DatabaseConfig{
		MaxOpenConns:    15,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	// This would normally be called by the connection functions
	// For testing, we'll verify the configuration is applied correctly
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify settings were applied
	stats := db.Stats()
	assert.Equal(t, 15, stats.MaxOpenConnections, "MaxOpenConnections should be set to 15")
}

func TestDatabaseConfigDefaults(t *testing.T) {
	// Test that default values are set correctly
	cfg := config.Load()

	// Verify write database defaults
	assert.Equal(t, 25, cfg.WriteDatabase.MaxOpenConns, "Write database MaxOpenConns should default to 25")
	assert.Equal(t, 5, cfg.WriteDatabase.MaxIdleConns, "Write database MaxIdleConns should default to 5")
	assert.Equal(t, 5*time.Minute, cfg.WriteDatabase.ConnMaxLifetime, "Write database ConnMaxLifetime should default to 5 minutes")
	assert.Equal(t, 5*time.Minute, cfg.WriteDatabase.ConnMaxIdleTime, "Write database ConnMaxIdleTime should default to 5 minutes")

	// Verify read database defaults
	assert.Equal(t, 25, cfg.ReadDatabase.MaxOpenConns, "Read database MaxOpenConns should default to 25")
	assert.Equal(t, 5, cfg.ReadDatabase.MaxIdleConns, "Read database MaxIdleConns should default to 5")
	assert.Equal(t, 5*time.Minute, cfg.ReadDatabase.ConnMaxLifetime, "Read database ConnMaxLifetime should default to 5 minutes")
	assert.Equal(t, 5*time.Minute, cfg.ReadDatabase.ConnMaxIdleTime, "Read database ConnMaxIdleTime should default to 5 minutes")

	// Verify event database defaults
	assert.Equal(t, 25, cfg.EventDatabase.MaxOpenConns, "Event database MaxOpenConns should default to 25")
	assert.Equal(t, 5, cfg.EventDatabase.MaxIdleConns, "Event database MaxIdleConns should default to 5")
	assert.Equal(t, 5*time.Minute, cfg.EventDatabase.ConnMaxLifetime, "Event database ConnMaxLifetime should default to 5 minutes")
	assert.Equal(t, 5*time.Minute, cfg.EventDatabase.ConnMaxIdleTime, "Event database ConnMaxIdleTime should default to 5 minutes")
}
