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
}

func New(cfg *config.Config) *Handler {
	// Create reverse proxy for order creation service
	orderCreationURL, _ := url.Parse(cfg.Proxy.OrderCreationURL)
	orderCreationProxy := httputil.NewSingleHostReverseProxy(orderCreationURL)

	// Configure proxy with timeout
	orderCreationProxy.Transport = &http.Transport{
		ResponseHeaderTimeout: cfg.Proxy.Timeout,
	}

	// Configure proxy to handle CORS headers properly
	orderCreationProxy.ModifyResponse = func(resp *http.Response) error {
		// Remove CORS headers from backend response since API Gateway handles them
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
	// Handle CORS preflight
	if c.Request.Method == "OPTIONS" {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(204)
		return
	}

	// Add proxy headers
	c.Request.Header.Set("X-Forwarded-By", "api-gateway")
	c.Request.Header.Set("X-Request-ID", c.GetString("RequestID"))

	// Create a custom response writer to handle duplicates
	writer := &cleanResponseWriter{
		ResponseWriter: c.Writer,
		headers:        make(map[string]string),
	}

	// Set CORS headers only once
	writer.headers["Access-Control-Allow-Origin"] = "*"
	writer.headers["Access-Control-Allow-Credentials"] = "true"

	// Proxy the request
	h.orderCreationProxy.ServeHTTP(writer, c.Request)
}

// cleanResponseWriter prevents duplicate headers
type cleanResponseWriter struct {
	http.ResponseWriter
	headers map[string]string
	written bool
}

func (w *cleanResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *cleanResponseWriter) WriteHeader(statusCode int) {
	if !w.written {
		// Set our clean headers
		for key, value := range w.headers {
			w.ResponseWriter.Header().Set(key, value)
		}
		w.written = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *cleanResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(200)
	}
	return w.ResponseWriter.Write(data)
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