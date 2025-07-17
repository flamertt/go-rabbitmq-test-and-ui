package app

import (
	"log"

	"go-rabbitmq-order-system/order-status-service/internal/config"
	"go-rabbitmq-order-system/order-status-service/internal/service"
	"go-rabbitmq-order-system/shared"
)

type App struct {
	config   *config.Config
	database *shared.Database
	rabbitMQ *shared.RabbitMQ
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Run() error {
	// Initialize database
	db, err := shared.NewDatabase(a.config.Database.URL)
	if err != nil {
		return err
	}
	a.database = db

	// Initialize RabbitMQ
	rabbitmq, err := shared.NewRabbitMQ(a.config.RabbitMQ.URL)
	if err != nil {
		return err
	}
	a.rabbitMQ = rabbitmq

	// Initialize service
	orderStatusService := service.New(db.DB, rabbitmq, &a.config.OrderStatus)

	// Start consuming events
	err = rabbitmq.ConsumeEvents("order_status_queue", orderStatusService.HandleOrderEvent)
	if err != nil {
		return err
	}

	log.Println("Order Status Update Service started")
	log.Println("Waiting for order events...")

	// Keep the service running
	select {}
}

func (a *App) Close() error {
	if a.database != nil {
		a.database.Close()
	}
	if a.rabbitMQ != nil {
		a.rabbitMQ.Close()
	}
	return nil
} 