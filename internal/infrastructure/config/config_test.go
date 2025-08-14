package config_test

import (
	"os"
	"testing"

	"go-clean-ddd-es-template/internal/infrastructure/config"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Test with default values
	cfg := config.Load()
	assert.NotNil(t, cfg)

	// Test write database config
	assert.Equal(t, "postgres", cfg.WriteDatabase.Type)
	assert.Equal(t, "localhost", cfg.WriteDatabase.Host)
	assert.Equal(t, "5432", cfg.WriteDatabase.Port)
	assert.Equal(t, "postgres", cfg.WriteDatabase.User)
	assert.Equal(t, "password", cfg.WriteDatabase.Password)
	assert.Equal(t, "clean_ddd_write_db", cfg.WriteDatabase.DBName)

	// Test message broker config
	assert.Equal(t, "kafka", cfg.MessageBroker.Type)
	assert.Equal(t, []string{"localhost:9092"}, cfg.MessageBroker.Brokers)
	assert.Equal(t, "user-events", cfg.MessageBroker.Topics["user.created"])
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("WRITE_DB_TYPE", "mysql")
	os.Setenv("WRITE_DB_HOST", "test-host")
	os.Setenv("WRITE_DB_PORT", "3306")
	os.Setenv("WRITE_DB_USER", "test-user")
	os.Setenv("WRITE_DB_PASSWORD", "test-pass")
	os.Setenv("WRITE_DB_NAME", "test-db")
	os.Setenv("MESSAGE_BROKER_TYPE", "rabbitmq")
	os.Setenv("MESSAGE_BROKER_BROKERS", "test-broker:5672")
	os.Setenv("MESSAGE_BROKER_TOPIC", "test-topic")

	defer func() {
		os.Unsetenv("WRITE_DB_TYPE")
		os.Unsetenv("WRITE_DB_HOST")
		os.Unsetenv("WRITE_DB_PORT")
		os.Unsetenv("WRITE_DB_USER")
		os.Unsetenv("WRITE_DB_PASSWORD")
		os.Unsetenv("WRITE_DB_NAME")
		os.Unsetenv("MESSAGE_BROKER_TYPE")
		os.Unsetenv("MESSAGE_BROKER_BROKERS")
		os.Unsetenv("MESSAGE_BROKER_TOPIC")
	}()

	cfg := config.Load()
	assert.NotNil(t, cfg)

	// Test write database config
	assert.Equal(t, "mysql", cfg.WriteDatabase.Type)
	assert.Equal(t, "test-host", cfg.WriteDatabase.Host)
	assert.Equal(t, "3306", cfg.WriteDatabase.Port)
	assert.Equal(t, "test-user", cfg.WriteDatabase.User)
	assert.Equal(t, "test-pass", cfg.WriteDatabase.Password)
	assert.Equal(t, "test-db", cfg.WriteDatabase.DBName)

	// Test message broker config
	assert.Equal(t, "rabbitmq", cfg.MessageBroker.Type)
	assert.Equal(t, []string{"test-broker:5672"}, cfg.MessageBroker.Brokers)
}

func TestDatabaseConfig_Fields(t *testing.T) {
	dbConfig := config.DatabaseConfig{
		Type:       "postgres",
		Host:       "localhost",
		Port:       "5432",
		User:       "postgres",
		Password:   "password",
		DBName:     "testdb",
		Collection: "users",
		Charset:    "utf8mb4",
		ParseTime:  true,
		Loc:        "Local",
	}

	assert.Equal(t, "postgres", dbConfig.Type)
	assert.Equal(t, "localhost", dbConfig.Host)
	assert.Equal(t, "5432", dbConfig.Port)
	assert.Equal(t, "postgres", dbConfig.User)
	assert.Equal(t, "password", dbConfig.Password)
	assert.Equal(t, "testdb", dbConfig.DBName)
	assert.Equal(t, "users", dbConfig.Collection)
	assert.Equal(t, "utf8mb4", dbConfig.Charset)
	assert.True(t, dbConfig.ParseTime)
	assert.Equal(t, "Local", dbConfig.Loc)
}

func TestMessageBrokerConfig_Fields(t *testing.T) {
	mbConfig := config.MessageBrokerConfig{
		Type:    "kafka",
		Brokers: []string{"localhost:9092"},
		Topics: map[string]string{
			"user.created": "user-events",
		},
		GroupID:  "user-service",
		Exchange: "user-events",
		Queue:    "user-events",
		Channel:  "user-events",
		Subject:  "user.events",
	}

	assert.Equal(t, "kafka", mbConfig.Type)
	assert.Equal(t, []string{"localhost:9092"}, mbConfig.Brokers)
	assert.Equal(t, "user-events", mbConfig.Topics["user.created"])
	assert.Equal(t, "user-service", mbConfig.GroupID)
	assert.Equal(t, "user-events", mbConfig.Exchange)
	assert.Equal(t, "user-events", mbConfig.Queue)
	assert.Equal(t, "user-events", mbConfig.Channel)
	assert.Equal(t, "user.events", mbConfig.Subject)
}

func TestTracingConfig_Fields(t *testing.T) {
	tracingConfig := config.TracingConfig{
		Enabled:     true,
		ServiceName: "test-service",
		Endpoint:    "http://localhost:4318/v1/traces",
	}

	assert.True(t, tracingConfig.Enabled)
	assert.Equal(t, "test-service", tracingConfig.ServiceName)
	assert.Equal(t, "http://localhost:4318/v1/traces", tracingConfig.Endpoint)
}

func TestServerConfig_Fields(t *testing.T) {
	serverConfig := config.ServerConfig{
		Port: "8080",
	}

	assert.Equal(t, "8080", serverConfig.Port)
}
