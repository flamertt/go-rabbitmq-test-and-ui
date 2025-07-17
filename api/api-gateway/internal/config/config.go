package config

import (
	"go-rabbitmq-order-system/pkg/config"
	"os"
	"strconv"
	"time"
)

type Config struct {
	*config.BaseConfig
	Server    ServerConfig
	RateLimit RateLimitConfig
	Proxy     ProxyConfig
	Auth      AuthConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
	Enabled           bool
}

type ProxyConfig struct {
	OrderCreationURL  string
	PaymentURL        string
	StockURL          string
	ShippingURL       string
	OrderStatusURL    string
	Timeout           time.Duration
}

type AuthConfig struct {
	AdminSecret  string
	JWTSecret    string
	TokenExpiry  time.Duration
	RequireAuth  bool
}

func Load() *Config {
	baseConfig := config.LoadBaseConfig()
	
	return &Config{
		BaseConfig: baseConfig,
		Server: ServerConfig{
			Port:         getEnv("GATEWAY_PORT", "8080"),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", "10s"),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", "10s"),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", "60s"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getEnvAsInt("RATE_LIMIT_RPS", 10),
			BurstSize:         getEnvAsInt("RATE_LIMIT_BURST", 20),
			Enabled:           getEnvAsBool("RATE_LIMIT_ENABLED", true),
		},
		Proxy: ProxyConfig{
			OrderCreationURL: getEnv("ORDER_CREATION_URL", "http://order-creation-service:8081"),
			PaymentURL:       getEnv("PAYMENT_URL", "http://payment-processing-service:8082"),
			StockURL:         getEnv("STOCK_URL", "http://stock-reservation-service:8083"),
			ShippingURL:      getEnv("SHIPPING_URL", "http://shipping-service:8084"),
			OrderStatusURL:   getEnv("ORDER_STATUS_URL", "http://order-status-service:8085"),
			Timeout:          getEnvAsDuration("PROXY_TIMEOUT", "30s"),
		},
		Auth: AuthConfig{
			AdminSecret:  getEnv("ADMIN_SECRET", "admin123"),
			JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
			TokenExpiry:  getEnvAsDuration("TOKEN_EXPIRY", "24h"),
			RequireAuth:  getEnvAsBool("REQUIRE_AUTH", false),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Minute
} 