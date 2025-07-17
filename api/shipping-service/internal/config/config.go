package config

import (
	"go-rabbitmq-order-system/pkg/config"
)

type Config struct {
	*config.BaseConfig
	Shipping ShippingConfig
}

type ShippingConfig struct {
	Carriers               []string
	ProcessingDelayMinutes int
	PremiumThreshold       float64
	StandardThreshold      float64
}

func Load() *Config {
	baseConfig := config.LoadBaseConfig()
	
	return &Config{
		BaseConfig: baseConfig,
		Shipping: ShippingConfig{
			Carriers: []string{
				"DHL", "UPS", "FedEx", 
				"Aras Kargo", "Yurtiçi Kargo", "PTT Kargo",
			},
			ProcessingDelayMinutes: 2,     // 2 minutes processing time
			PremiumThreshold:       5000,  // Premium shipping for orders > 5000
			StandardThreshold:      1000,  // Standard shipping for orders > 1000
		},
	}
} 