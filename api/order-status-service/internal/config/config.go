package config

import (
	"go-rabbitmq-order-system/pkg/config"
)

type Config struct {
	*config.BaseConfig
	OrderStatus OrderStatusConfig
}

type OrderStatusConfig struct {
	UpdateBatchSize int
	LogLevel        string
	EnableAuditLog  bool
}

func Load() *Config {
	baseConfig := config.LoadBaseConfig()
	
	return &Config{
		BaseConfig: baseConfig,
		OrderStatus: OrderStatusConfig{
			UpdateBatchSize: 100,   // Process 100 updates at a time
			LogLevel:        "INFO", // INFO, DEBUG, WARN, ERROR
			EnableAuditLog:  true,   // Enable audit logging
		},
	}
} 