package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"go-rabbitmq-order-system/auth-service/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	authResponse, err := h.authService.Register(&req)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			h.sendError(w, http.StatusConflict, "User already exists", err.Error())
		default:
			if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "required") {
				h.sendError(w, http.StatusBadRequest, "Validation error", err.Error())
			} else {
				h.sendError(w, http.StatusInternalServerError, "Registration failed", err.Error())
			}
		}
		return
	}

	h.sendSuccess(w, http.StatusCreated, authResponse, "User registered successfully")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ipAddress := h.getClientIP(r)

	authResponse, err := h.authService.Login(&req, userAgent, ipAddress)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			h.sendError(w, http.StatusUnauthorized, "Invalid credentials", err.Error())
		case service.ErrUserInactive:
			h.sendError(w, http.StatusForbidden, "Account inactive", err.Error())
		default:
			if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "required") {
				h.sendError(w, http.StatusBadRequest, "Validation error", err.Error())
			} else {
				h.sendError(w, http.StatusInternalServerError, "Login failed", err.Error())
			}
		}
		return
	}

	h.sendSuccess(w, http.StatusOK, authResponse, "Login successful")
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.RefreshToken == "" {
		h.sendError(w, http.StatusBadRequest, "Refresh token required", "")
		return
	}

	authResponse, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrInvalidToken, service.ErrTokenExpired:
			h.sendError(w, http.StatusUnauthorized, "Invalid or expired token", err.Error())
		case service.ErrUserInactive:
			h.sendError(w, http.StatusForbidden, "Account inactive", err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, "Token refresh failed", err.Error())
		}
		return
	}

	h.sendSuccess(w, http.StatusOK, authResponse, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	token := h.extractTokenFromHeader(r)
	if token == "" {
		h.sendError(w, http.StatusBadRequest, "Authorization token required", "")
		return
	}

	if err := h.authService.Logout(token); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Logout failed", err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, nil, "Logout successful")
}

func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	token := h.extractTokenFromHeader(r)
	if token == "" {
		h.sendError(w, http.StatusBadRequest, "Authorization token required", "")
		return
	}

	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		switch err {
		case service.ErrInvalidToken, service.ErrTokenExpired:
			h.sendError(w, http.StatusUnauthorized, "Invalid or expired token", err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, "Token validation failed", err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"valid":   true,
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
	}

	h.sendSuccess(w, http.StatusOK, response, "Token is valid")
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	token := h.extractTokenFromHeader(r)
	if token == "" {
		h.sendError(w, http.StatusBadRequest, "Authorization token required", "")
		return
	}

	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		switch err {
		case service.ErrInvalidToken, service.ErrTokenExpired:
			h.sendError(w, http.StatusUnauthorized, "Invalid or expired token", err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, "Token validation failed", err.Error())
		}
		return
	}

	user, err := h.authService.GetUserProfile(claims.UserID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get user profile", err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, user, "Profile retrieved successfully")
}

func (h *AuthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	response := map[string]interface{}{
		"service": "auth-service",
		"status":  "healthy",
		"version": "1.0.0",
	}

	h.sendSuccess(w, http.StatusOK, response, "Service is healthy")
}

// Helper methods
func (h *AuthHandler) sendError(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   error,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) sendSuccess(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Bearer token format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func (h *AuthHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs separated by commas
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
} 