package auth

import (
	"time"

	"QuanPhotos/internal/pkg/jwt"
	"QuanPhotos/internal/repository/postgresql/token"
	"QuanPhotos/internal/repository/postgresql/user"
)

// Service handles authentication business logic
type Service struct {
	userRepo   *user.UserRepository
	tokenRepo  *token.TokenRepository
	jwtManager *jwt.Manager
}

// New creates a new auth service
func New(userRepo *user.UserRepository, tokenRepo *token.TokenRepository, jwtManager *jwt.Manager) *Service {
	return &Service{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
	}
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// RegisterRequest represents registration request data
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest represents login request data
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
