package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Kafka    KafkaConfig    `json:"kafka"`
	Logger   LoggerConfig   `json:"logger"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Database     string `json:"database"`
	SSLMode      string `json:"ssl_mode"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers           []string `json:"brokers"`
	TransactionsTopic string   `json:"transactions_topic"`
	ConsumerGroup     string   `json:"consumer_group"`
	BatchSize         int      `json:"batch_size"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getIntEnv("DB_PORT", 5432),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "password"),
			Database:     getEnv("DB_NAME", "ledgertime"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 25),
		},
		Kafka: KafkaConfig{
			Brokers:           getSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
			TransactionsTopic: getEnv("KAFKA_TRANSACTIONS_TOPIC", "card-transactions"),
			ConsumerGroup:     getEnv("KAFKA_CONSUMER_GROUP", "ledger-consumer"),
			BatchSize:         getIntEnv("KAFKA_BATCH_SIZE", 100),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return cfg, nil
}

// GetDSN returns the database connection string
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		return []string{value}
	}
	return defaultValue
}
