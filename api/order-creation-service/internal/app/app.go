package app

import (
	"log"
	"net/http"

	"go-rabbitmq-order-system/order-creation-service/internal/config"
	"go-rabbitmq-order-system/order-creation-service/internal/handler"
	"go-rabbitmq-order-system/order-creation-service/internal/repository"
	"go-rabbitmq-order-system/order-creation-service/internal/service"
	"go-rabbitmq-order-system/pkg/middleware"
	"go-rabbitmq-order-system/shared"

	"github.com/gin-gonic/gin"
)

type App struct {
	config   *config.Config
	router   *gin.Engine
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

	// Setup RabbitMQ exchanges and queues
	if err := rabbitmq.SetupExchangeAndQueues(); err != nil {
		return err
	}

	// Initialize layers
	repo := repository.New(db.DB)
	svc := service.New(repo, rabbitmq)
	h := handler.New(svc)

	// Setup router
	a.setupRouter(h)

	log.Printf("Order Creation Service started on port %s", a.config.Server.Port)
	return http.ListenAndServe(":"+a.config.Server.Port, a.router)
}

func (a *App) setupRouter(h *handler.Handler) {
	r := gin.Default()

	// Add middleware (skip CORS since API Gateway handles it)
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger("order-creation"))
	r.Use(middleware.Recovery("order-creation"))

	// Health check
	r.GET("/health", h.Health)

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/orders", h.GetOrders)
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders/:id", h.GetOrder)
		api.GET("/products", h.GetProducts)
		api.GET("/products/:id", h.GetProduct)
	}

	a.router = r
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