package config

import (
	"os"
	"strconv"
	"time"
)

// BaseConfig contains common configuration for all services
type BaseConfig struct {
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	Redis    RedisConfig
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RabbitMQConfig struct {
	URL              string
	ExchangeName     string
	RetryAttempts    int
	RetryDelay       time.Duration
	HeartbeatTimeout time.Duration
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// LoadBaseConfig loads common configuration for all services
func LoadBaseConfig() *BaseConfig {
	return &BaseConfig{
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://admin:admin123@localhost:5432/order_system?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", "5m"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:              getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			ExchangeName:     getEnv("RABBITMQ_EXCHANGE", "order_events_exchange"),
			RetryAttempts:    getEnvAsInt("RABBITMQ_RETRY_ATTEMPTS", 3),
			RetryDelay:       getEnvAsDuration("RABBITMQ_RETRY_DELAY", "1s"),
			HeartbeatTimeout: getEnvAsDuration("RABBITMQ_HEARTBEAT", "10s"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
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

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// If parsing fails, parse the default
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Minute // fallback
} 