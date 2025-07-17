package main

import (
	"log"

	"go-rabbitmq-order-system/shipping-service/internal/app"
	"go-rabbitmq-order-system/shipping-service/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize and run application
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		log.Fatal("Failed to run application:", err)
	}
} 