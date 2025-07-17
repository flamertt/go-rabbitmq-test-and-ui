package config

import (
	"go-rabbitmq-order-system/pkg/config"
)

type Config struct {
	*config.BaseConfig
	StockReservation StockReservationConfig
}

type StockReservationConfig struct {
	ReservationTimeoutMinutes int
	LockTimeoutSeconds        int
	RetryAttempts             int
}

func Load() *Config {
	baseConfig := config.LoadBaseConfig()
	
	return &Config{
		BaseConfig: baseConfig,
		StockReservation: StockReservationConfig{
			ReservationTimeoutMinutes: 15, // 15 minutes reservation timeout
			LockTimeoutSeconds:        30, // 30 seconds lock timeout
			RetryAttempts:             3,  // 3 retry attempts
		},
	}
} 