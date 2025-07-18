package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"go-rabbitmq-order-system/auth-service/internal/config"
	"go-rabbitmq-order-system/auth-service/internal/handler"
	"go-rabbitmq-order-system/auth-service/internal/service"

	_ "github.com/lib/pq"
)

type App struct {
	config  *config.Config
	db      *sql.DB
	handler *handler.AuthHandler
}

func New(cfg *config.Config) (*App, error) {
	// Initialize database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	// Initialize services
	authService := service.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpiration, cfg.RefreshExpiration, cfg.BCryptCost)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)

	return &App{
		config:  cfg,
		db:      db,
		handler: authHandler,
	}, nil
}

func (a *App) Run() error {
	// Setup routes
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/auth/register", a.handler.Register)
	mux.HandleFunc("/auth/login", a.handler.Login)
	mux.HandleFunc("/auth/refresh", a.handler.RefreshToken)
	mux.HandleFunc("/auth/logout", a.handler.Logout)
	mux.HandleFunc("/auth/validate", a.handler.ValidateToken)
	mux.HandleFunc("/auth/profile", a.handler.GetProfile)

	// Health check
	mux.HandleFunc("/health", a.handler.HealthCheck)

	// Add CORS middleware
	handler := a.corsMiddleware(mux)

	log.Printf("Auth Service starting on port %s", a.config.Port)
	log.Println("Available endpoints:")
	log.Println("  POST /auth/register")
	log.Println("  POST /auth/login")
	log.Println("  POST /auth/refresh")
	log.Println("  POST /auth/logout")
	log.Println("  POST /auth/validate")
	log.Println("  GET  /auth/profile")
	log.Println("  GET  /health")

	return http.ListenAndServe(":"+a.config.Port, handler)
}

func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// CORS middleware
func (a *App) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
} 