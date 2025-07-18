package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	authServiceURL string
}

type AuthResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Valid  bool   `json:"valid"`
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	} `json:"data"`
}

func NewAuthMiddleware(authServiceURL string) *AuthMiddleware {
	return &AuthMiddleware{
		authServiceURL: authServiceURL,
	}
}

func (am *AuthMiddleware) ValidateToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Bearer token format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token with auth service
		if !am.validateTokenWithAuthService(token) {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to next handler
		next.ServeHTTP(w, r)
	}
}

func (am *AuthMiddleware) validateTokenWithAuthService(token string) bool {
	client := &http.Client{}
	
	req, err := http.NewRequest("POST", am.authServiceURL+"/auth/validate", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return false
	}

	return authResp.Success && authResp.Data.Valid
} 