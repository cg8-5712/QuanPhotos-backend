package auth

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/hash"
)

var (
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")
)

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*model.User, *TokenPair, error) {
	// Check if username exists
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, ErrUsernameExists
	}

	// Check if email exists
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, ErrEmailExists
	}

	// Hash password
	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	// Create user
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         model.RoleUser,
		Status:       model.StatusActive,
		CanComment:   true,
		CanMessage:   true,
		CanUpload:    true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Generate tokens
	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// generateTokenPair generates access and refresh tokens
func (s *Service) generateTokenPair(ctx context.Context, user *model.User) (*TokenPair, error) {
	// Generate access token
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Hash refresh token for storage
	tokenHash, err := hash.HashPassword(refreshToken)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: refreshExpiresAt,
	}

	if err := s.tokenRepo.CreateWithCleanup(ctx, rt, 5); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}
