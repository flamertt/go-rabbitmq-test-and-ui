package app

import (
	"log"
	"net/http"

	"go-rabbitmq-order-system/api-gateway/internal/config"
	"go-rabbitmq-order-system/api-gateway/internal/handler"
	"go-rabbitmq-order-system/api-gateway/internal/middleware"
	commonMiddleware "go-rabbitmq-order-system/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type App struct {
	config *config.Config
	router *gin.Engine
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Run() error {
	// Initialize handler
	h := handler.New(a.config)

	// Setup router
	a.setupRouter(h)

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + a.config.Server.Port,
		Handler:      a.router,
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
		IdleTimeout:  a.config.Server.IdleTimeout,
	}

	log.Printf("API Gateway started on port %s", a.config.Server.Port)
	return server.ListenAndServe()
}

func (a *App) setupRouter(h *handler.Handler) {
	r := gin.Default()

	// Other middleware (NO CORS here, handled in proxy)
	r.Use(commonMiddleware.RequestID())
	r.Use(commonMiddleware.StructuredLogger("api-gateway"))
	r.Use(commonMiddleware.Recovery("api-gateway"))

	// Rate limiting middleware
	if a.config.RateLimit.Enabled {
		r.Use(middleware.RateLimit(a.config.RateLimit.RequestsPerSecond, a.config.RateLimit.BurstSize))
	}

	// Health check
	r.GET("/health", h.Health)

	// Admin routes
	admin := r.Group("/admin")
	if a.config.Auth.RequireAuth {
		admin.Use(middleware.AdminAuth(a.config.Auth.AdminSecret))
	}
	{
		admin.GET("/status", h.AdminStatus)
		admin.GET("/metrics", h.Metrics)
	}

	// API routes with proxy
	api := r.Group("/api/v1")
	{
		// Order Creation Service routes
		orders := api.Group("/orders")
		{
			orders.OPTIONS("", h.ProxyToOrderCreation)        // preflight for POST /orders
			orders.OPTIONS("/:id", h.ProxyToOrderCreation)   // preflight for GET /orders/:id
			orders.GET("", h.ProxyToOrderCreation)           // GET all orders
			orders.POST("", h.ProxyToOrderCreation)
			orders.GET("/:id", h.ProxyToOrderCreation)
		}

		// Product routes
		api.OPTIONS("/products", h.ProxyToOrderCreation)     // preflight for GET /products
		api.OPTIONS("/products/:id", h.ProxyToOrderCreation) // preflight for GET /products/:id
		api.GET("/products", h.ProxyToOrderCreation)
		api.GET("/products/:id", h.ProxyToOrderCreation)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.OPTIONS("/register", h.AuthRegister)
			auth.OPTIONS("/login", h.AuthLogin) 
			auth.OPTIONS("/refresh", h.AuthRefresh)
			auth.OPTIONS("/logout", h.AuthLogout)
			auth.OPTIONS("/validate", h.AuthValidate)
			auth.OPTIONS("/profile", h.AuthProfile)
			
			auth.POST("/register", h.AuthRegister)
			auth.POST("/login", h.AuthLogin)
			auth.POST("/refresh", h.AuthRefresh)
			auth.POST("/logout", h.AuthLogout)
			auth.POST("/validate", h.AuthValidate)
			auth.GET("/profile", h.AuthProfile)
		}
	}

	a.router = r
} 