package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"go-rabbitmq-order-system/api-gateway/internal/config"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config             *config.Config
	orderCreationProxy *httputil.ReverseProxy
	authServiceProxy   *httputil.ReverseProxy
}

func New(cfg *config.Config) *Handler {
	// Create reverse proxy for order creation service
	orderCreationURL, _ := url.Parse(cfg.Proxy.OrderCreationURL)
	orderCreationProxy := httputil.NewSingleHostReverseProxy(orderCreationURL)

	// Create reverse proxy for auth service
	authServiceURL, _ := url.Parse(cfg.Proxy.AuthServiceURL)
	authServiceProxy := httputil.NewSingleHostReverseProxy(authServiceURL)

	// Configure proxy with timeout
	orderCreationProxy.Transport = &http.Transport{
		ResponseHeaderTimeout: cfg.Proxy.Timeout,
	}

	authServiceProxy.Transport = &http.Transport{
		ResponseHeaderTimeout: cfg.Proxy.Timeout,
	}

	// Configure proxy to handle headers properly
	orderCreationProxy.ModifyResponse = func(resp *http.Response) error {
		// Remove any CORS headers from backend to prevent duplicates
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Max-Age")
		
		// Add test header to verify ModifyResponse works
		resp.Header.Set("X-Proxy-Modified", "true")
		return nil
	}

	authServiceProxy.ModifyResponse = func(resp *http.Response) error {
		// Remove any CORS headers from backend to prevent duplicates
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Max-Age")
		
		// Add test header to verify ModifyResponse works
		resp.Header.Set("X-Proxy-Modified", "true")
		return nil
	}

	return &Handler{
		config:             cfg,
		orderCreationProxy: orderCreationProxy,
		authServiceProxy:   authServiceProxy,
	}
}

func (h *Handler) Health(c *gin.Context) {
	// Set CORS headers
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Credentials", "true")
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "api-gateway",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

func (h *Handler) AdminStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"services": gin.H{
			"order-creation": h.checkServiceHealth(h.config.Proxy.OrderCreationURL),
			"payment":        "unknown", // TODO: implement health checks
			"stock":          "unknown",
			"shipping":       "unknown",
			"order-status":   "unknown",
		},
		"config": gin.H{
			"rate_limit_enabled": h.config.RateLimit.Enabled,
			"rate_limit_rps":     h.config.RateLimit.RequestsPerSecond,
			"auth_required":      h.config.Auth.RequireAuth,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) Metrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"metrics": gin.H{
			"total_requests":  "TODO", // TODO: implement metrics collection
			"active_requests": "TODO",
			"error_rate":      "TODO",
			"avg_response_time": "TODO",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) ProxyToOrderCreation(c *gin.Context) {
	// Set CORS headers
	h.setCORSHeaders(c)
	
	// Handle CORS preflight
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}

	// Add proxy headers
	c.Request.Header.Set("X-Forwarded-By", "api-gateway")
	c.Request.Header.Set("X-Request-ID", c.GetString("RequestID"))

	// Proxy the request directly
	h.orderCreationProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) checkServiceHealth(serviceURL string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(serviceURL + "/health")
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "healthy"
	}

	return "unhealthy"
}

func (h *Handler) setCORSHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
	c.Header("Access-Control-Max-Age", "86400")
}

// Auth service proxy methods
func (h *Handler) AuthRegister(c *gin.Context) {
	h.setCORSHeaders(c)
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}
	// Rewrite path from /api/v1/auth/register to /auth/register
	c.Request.URL.Path = "/auth/register"
	h.authServiceProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) AuthLogin(c *gin.Context) {
	h.setCORSHeaders(c)
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}
	// Rewrite path from /api/v1/auth/login to /auth/login
	c.Request.URL.Path = "/auth/login"
	h.authServiceProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) AuthRefresh(c *gin.Context) {
	h.setCORSHeaders(c)
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}
	// Rewrite path from /api/v1/auth/refresh to /auth/refresh
	c.Request.URL.Path = "/auth/refresh"
	h.authServiceProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) AuthLogout(c *gin.Context) {
	h.setCORSHeaders(c)
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}
	// Rewrite path from /api/v1/auth/logout to /auth/logout
	c.Request.URL.Path = "/auth/logout"
	h.authServiceProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) AuthValidate(c *gin.Context) {
	h.setCORSHeaders(c)
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}
	// Rewrite path from /api/v1/auth/validate to /auth/validate
	c.Request.URL.Path = "/auth/validate"
	h.authServiceProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) AuthProfile(c *gin.Context) {
	h.setCORSHeaders(c)
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}
	// Rewrite path from /api/v1/auth/profile to /auth/profile
	c.Request.URL.Path = "/auth/profile"
	h.authServiceProxy.ServeHTTP(c.Writer, c.Request)
} 