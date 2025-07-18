package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiration  time.Duration
	RefreshExpiration time.Duration
	BCryptCost     int
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://orderuser:orderpass123@localhost:5432/order_system?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "order-system-jwt-secret-key-2024-very-secure"
	}

	jwtExpirationStr := os.Getenv("JWT_EXPIRATION_HOURS")
	jwtExpiration := 24 * time.Hour // default 24 hours
	if jwtExpirationStr != "" {
		if hours, err := strconv.Atoi(jwtExpirationStr); err == nil {
			jwtExpiration = time.Duration(hours) * time.Hour
		}
	}

	refreshExpirationStr := os.Getenv("REFRESH_EXPIRATION_DAYS")
	refreshExpiration := 7 * 24 * time.Hour // default 7 days
	if refreshExpirationStr != "" {
		if days, err := strconv.Atoi(refreshExpirationStr); err == nil {
			refreshExpiration = time.Duration(days) * 24 * time.Hour
		}
	}

	bcryptCostStr := os.Getenv("BCRYPT_COST")
	bcryptCost := 12 // default cost
	if bcryptCostStr != "" {
		if cost, err := strconv.Atoi(bcryptCostStr); err == nil && cost >= 4 && cost <= 31 {
			bcryptCost = cost
		}
	}

	return &Config{
		Port:              port,
		DatabaseURL:       databaseURL,
		JWTSecret:         jwtSecret,
		JWTExpiration:     jwtExpiration,
		RefreshExpiration: refreshExpiration,
		BCryptCost:        bcryptCost,
	}, nil
} 