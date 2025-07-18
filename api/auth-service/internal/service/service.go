package service

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db                *sql.DB
	jwtSecret         []byte
	jwtExpiration     time.Duration
	refreshExpiration time.Duration
	bcryptCost        int
}

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUserInactive       = errors.New("user account is inactive")
)

func NewAuthService(db *sql.DB, jwtSecret string, jwtExpiration, refreshExpiration time.Duration, bcryptCost int) *AuthService {
	return &AuthService{
		db:                db,
		jwtSecret:         []byte(jwtSecret),
		jwtExpiration:     jwtExpiration,
		refreshExpiration: refreshExpiration,
		bcryptCost:        bcryptCost,
	}
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID := uuid.New().String()
	now := time.Now()

	_, err = s.db.Exec(`
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_active, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		userID, strings.ToLower(req.Email), string(hashedPassword), req.FirstName, req.LastName, "customer", true, false, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get created user
	user, err := s.getUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) Login(req *LoginRequest, userAgent, ipAddress string) (*AuthResponse, error) {
	// Validate input
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// Get user by email
	user, passwordHash, err := s.getUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Store session
	if err := s.storeSession(user.ID, accessToken, refreshToken, userAgent, ipAddress); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (s *AuthService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.getUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Update session
	if err := s.updateSession(claims.UserID, newAccessToken, newRefreshToken); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) Logout(tokenString string) error {
	// Parse token to get user ID
	_, err := s.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	// Delete session
	tokenHash := s.hashToken(tokenString)
	_, err = s.db.Exec("DELETE FROM user_sessions WHERE token_hash = $1", tokenHash)
	return err
}

func (s *AuthService) GetUserProfile(userID string) (*User, error) {
	return s.getUserByID(userID)
}

// Helper methods
func (s *AuthService) validateRegisterRequest(req *RegisterRequest) error {
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return errors.New("all fields are required")
	}

	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	if !strings.Contains(req.Email, "@") {
		return errors.New("invalid email format")
	}

	return nil
}

func (s *AuthService) validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return errors.New("email and password are required")
	}
	return nil
}

func (s *AuthService) getUserByEmail(email string) (*User, string, error) {
	var user User
	var passwordHash string

	err := s.db.QueryRow(`
		SELECT id, email, password_hash, first_name, last_name, role, is_active, email_verified, created_at, updated_at
		FROM users WHERE email = $1`, strings.ToLower(email)).Scan(
		&user.ID, &user.Email, &passwordHash, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, "", err
	}

	return &user, passwordHash, nil
}

func (s *AuthService) getUserByID(userID string) (*User, error) {
	var user User

	err := s.db.QueryRow(`
		SELECT id, email, first_name, last_name, role, is_active, email_verified, created_at, updated_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) generateTokens(user *User) (string, string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(s.jwtExpiration)
	refreshExpiresAt := now.Add(s.refreshExpiration)

	// Access token claims
	accessClaims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	// Refresh token claims
	refreshClaims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	// Generate tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	return accessTokenString, refreshTokenString, expiresAt.Unix(), nil
}

func (s *AuthService) storeSession(userID, accessToken, refreshToken, userAgent, ipAddress string) error {
	sessionID := uuid.New().String()
	tokenHash := s.hashToken(accessToken)
	refreshTokenHash := s.hashToken(refreshToken)
	expiresAt := time.Now().Add(s.refreshExpiration)

	// Store IP address as string to avoid UTF8 encoding issues
	var ipStr *string
	if ipAddress != "" {
		ipStr = &ipAddress
	}

	_, err := s.db.Exec(`
		INSERT INTO user_sessions (id, user_id, token_hash, refresh_token_hash, expires_at, created_at, last_used_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		sessionID, userID, tokenHash, refreshTokenHash, expiresAt, time.Now(), time.Now(), userAgent, ipStr)

	return err
}

func (s *AuthService) updateSession(userID, accessToken, refreshToken string) error {
	tokenHash := s.hashToken(accessToken)
	refreshTokenHash := s.hashToken(refreshToken)
	expiresAt := time.Now().Add(s.refreshExpiration)

	_, err := s.db.Exec(`
		UPDATE user_sessions 
		SET token_hash = $1, refresh_token_hash = $2, expires_at = $3, last_used_at = $4
		WHERE user_id = $5`,
		tokenHash, refreshTokenHash, expiresAt, time.Now(), userID)

	return err
}

func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
} 