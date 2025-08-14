package config

import (
	"os"
	"strings"
)

type Config struct {
	Server        ServerConfig
	WriteDatabase DatabaseConfig
	ReadDatabase  DatabaseConfig
	EventDatabase DatabaseConfig
	MessageBroker MessageBrokerConfig
	Tracing       TracingConfig
	Log           LogConfig
	I18n          I18nConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Type     string // "postgres", "mysql", "mongodb"
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	// MongoDB specific
	URI        string
	Collection string
	// MySQL specific
	Charset   string
	ParseTime bool
	Loc       string
}

type MessageBrokerConfig struct {
	Type    string // "kafka", "rabbitmq", "redis", "nats"
	Brokers []string
	Topics  map[string]string
	// Kafka specific
	GroupID string
	// RabbitMQ specific
	Exchange string
	Queue    string
	// Redis specific
	Channel string
	// NATS specific
	Subject string
}

type TracingConfig struct {
	Enabled     bool
	ServiceName string
	Endpoint    string
}

type LogConfig struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "json", "text"
}

type I18nConfig struct {
	DefaultLocale   string
	TranslationsDir string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		WriteDatabase: DatabaseConfig{
			Type:       getEnv("WRITE_DB_TYPE", "postgres"),
			Host:       getEnv("WRITE_DB_HOST", "localhost"),
			Port:       getEnv("WRITE_DB_PORT", "5432"),
			User:       getEnv("WRITE_DB_USER", "postgres"),
			Password:   getEnv("WRITE_DB_PASSWORD", "password"),
			DBName:     getEnv("WRITE_DB_NAME", "clean_ddd_write_db"),
			Collection: getEnv("WRITE_DB_COLLECTION", "users"),
			Charset:    getEnv("WRITE_DB_CHARSET", "utf8mb4"),
			ParseTime:  getEnv("WRITE_DB_PARSE_TIME", "true") == "true",
			Loc:        getEnv("WRITE_DB_LOC", "Local"),
		},
		ReadDatabase: DatabaseConfig{
			Type:       getEnv("READ_DB_TYPE", "mongodb"),
			Host:       getEnv("READ_DB_HOST", "localhost"),
			Port:       getEnv("READ_DB_PORT", "27017"),
			User:       getEnv("READ_DB_USER", ""),
			Password:   getEnv("READ_DB_PASSWORD", ""),
			DBName:     getEnv("READ_DB_NAME", "clean_ddd_read_db"),
			URI:        getEnv("READ_DB_URI", "mongodb://localhost:27017"),
			Collection: getEnv("READ_DB_COLLECTION", "users"),
			Charset:    getEnv("READ_DB_CHARSET", "utf8mb4"),
			ParseTime:  getEnv("READ_DB_PARSE_TIME", "true") == "true",
			Loc:        getEnv("READ_DB_LOC", "Local"),
		},
		EventDatabase: DatabaseConfig{
			Type:       getEnv("EVENT_DB_TYPE", "postgres"),
			Host:       getEnv("EVENT_DB_HOST", "localhost"),
			Port:       getEnv("EVENT_DB_PORT", "5432"),
			User:       getEnv("EVENT_DB_USER", "postgres"),
			Password:   getEnv("EVENT_DB_PASSWORD", "password"),
			DBName:     getEnv("EVENT_DB_NAME", "clean_ddd_event_db"),
			Collection: getEnv("EVENT_DB_COLLECTION", "events"),
			Charset:    getEnv("EVENT_DB_CHARSET", "utf8mb4"),
			ParseTime:  getEnv("EVENT_DB_PARSE_TIME", "true") == "true",
			Loc:        getEnv("EVENT_DB_LOC", "Local"),
		},
		MessageBroker: MessageBrokerConfig{
			Type:    getEnv("MESSAGE_BROKER_TYPE", "kafka"),
			Brokers: strings.Split(getEnv("MESSAGE_BROKER_BROKERS", "localhost:9092"), ","),
			Topics: map[string]string{
				"user.created": "user-events",
				"user.updated": "user-events",
				"user.deleted": "user-events",
			},
			GroupID:  getEnv("MESSAGE_BROKER_GROUP_ID", "user-service"),
			Exchange: getEnv("MESSAGE_BROKER_EXCHANGE", "user-events"),
			Queue:    getEnv("MESSAGE_BROKER_QUEUE", "user-events"),
			Channel:  getEnv("MESSAGE_BROKER_CHANNEL", "user-events"),
			Subject:  getEnv("MESSAGE_BROKER_SUBJECT", "user.events"),
		},
		Tracing: TracingConfig{
			Enabled:     getEnv("TRACING_ENABLED", "true") == "true",
			ServiceName: getEnv("TRACING_SERVICE_NAME", "go-clean-ddd-es-template"),
			Endpoint:    getEnv("TRACING_ENDPOINT", "http://localhost:4318/v1/traces"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		I18n: I18nConfig{
			DefaultLocale:   getEnv("I18N_DEFAULT_LOCALE", "en"),
			TranslationsDir: getEnv("I18N_TRANSLATIONS_DIR", "./translations"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
