package config

import (
	"go-rabbitmq-order-system/pkg/config"
)

type Config struct {
	*config.BaseConfig
	PaymentGateway PaymentGatewayConfig
}

type PaymentGatewayConfig struct {
	SuccessRate      float64
	ProcessingDelayMS int
}

func Load() *Config {
	baseConfig := config.LoadBaseConfig()
	
	return &Config{
		BaseConfig: baseConfig,
		PaymentGateway: PaymentGatewayConfig{
			SuccessRate:      0.9, // 90% success rate
			ProcessingDelayMS: 2000, // 2 seconds
		},
	}
} 