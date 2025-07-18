package main

import (
	"log"
	"os"

	"go-rabbitmq-order-system/auth-service/internal/app"
	"go-rabbitmq-order-system/auth-service/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize and run application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
		os.Exit(1)
	}
} 