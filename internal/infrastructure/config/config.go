package config

import (
	"os"
	"strconv"
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
	Auth          AuthConfig
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
	Level      string `json:"level" yaml:"level"`             // "debug", "info", "warn", "error", "fatal"
	Format     string `json:"format" yaml:"format"`           // "json", "text", "console"
	Output     string `json:"output" yaml:"output"`           // "stdout", "stderr", "file"
	FilePath   string `json:"file_path" yaml:"file_path"`     // Path to log file when output is "file"
	MaxSize    int    `json:"max_size" yaml:"max_size"`       // Max size in MB for log rotation
	MaxBackups int    `json:"max_backups" yaml:"max_backups"` // Max number of backup files
	MaxAge     int    `json:"max_age" yaml:"max_age"`         // Max age in days for log files
	Compress   bool   `json:"compress" yaml:"compress"`       // Whether to compress rotated log files
	Caller     bool   `json:"caller" yaml:"caller"`           // Whether to include caller information
	Stacktrace bool   `json:"stacktrace" yaml:"stacktrace"`   // Whether to include stack trace for errors
}

type I18nConfig struct {
	DefaultLocale   string
	TranslationsDir string
}

type AuthConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	TokenExpiry    int // in hours
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
			User:       getEnv("READ_DB_USER", "admin"),
			Password:   getEnv("READ_DB_PASSWORD", "password"),
			DBName:     getEnv("READ_DB_NAME", "clean_ddd_read_db"),
			URI:        getEnv("READ_DB_URI", "mongodb://admin:password@localhost:27017"),
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
				// High-volume events: Separate topics
				"user.login":   "user.login",
				"user.logout":  "user.logout",
				"product.view": "product.view",

				// Medium-volume events: Domain-grouped topics
				"user.created":    "user-events",
				"user.updated":    "user-events",
				"user.deleted":    "user-events",
				"order.created":   "order-events",
				"order.updated":   "order-events",
				"order.cancelled": "order-events",
				"product.created": "product-events",
				"product.updated": "product-events",
				"product.deleted": "product-events",

				// Low-volume events: Bounded-context grouped topics
				"admin.login":        "admin-events",
				"admin.logout":       "admin-events",
				"admin.create_user":  "admin-events",
				"admin.delete_user":  "admin-events",
				"system.backup":      "system-events",
				"system.maintenance": "system-events",
				"audit.log":          "audit-events",
				"security.event":     "audit-events",
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
			Endpoint:    getEnv("TRACING_ENDPOINT", "localhost:4318"),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "text"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			FilePath:   getEnv("LOG_FILE_PATH", "./logs/app.log"),
			MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvAsInt("LOG_MAX_AGE", 28),
			Compress:   getEnv("LOG_COMPRESS", "true") == "true",
			Caller:     getEnv("LOG_CALLER", "true") == "true",
			Stacktrace: getEnv("LOG_STACKTRACE", "true") == "true",
		},
		I18n: I18nConfig{
			DefaultLocale:   getEnv("I18N_DEFAULT_LOCALE", "en"),
			TranslationsDir: getEnv("I18N_TRANSLATIONS_DIR", "./translations"),
		},
		Auth: AuthConfig{
			PrivateKeyPath: getEnv("AUTH_PRIVATE_KEY_PATH", "./keys/private.pem"),
			PublicKeyPath:  getEnv("AUTH_PUBLIC_KEY_PATH", "./keys/public.pem"),
			TokenExpiry:    getEnvAsInt("AUTH_TOKEN_EXPIRY", 24), // 24 hours
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
