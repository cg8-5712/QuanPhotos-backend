package auth

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/hash"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUsernameExists    = errors.New("username already exists")
	ErrEmailExists       = errors.New("email already exists")
	ErrPasswordMismatch  = errors.New("passwords do not match")
)

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*model.User, *TokenPair, error) {
	// Validate password confirmation
	if req.Password != req.ConfirmPassword {
		return nil, nil, ErrPasswordMismatch
	}

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

	// Start transaction to ensure atomicity
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	// Create user within transaction
	if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
		return nil, nil, err
	}

	// Generate tokens within transaction
	tokens, err := s.generateTokenPairTx(ctx, tx, user)
	if err != nil {
		return nil, nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
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

	// Hash refresh token for storage (use SHA-256, not bcrypt - JWT tokens exceed bcrypt's 72-byte limit)
	tokenHash := hash.HashToken(refreshToken)

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

// generateTokenPairTx generates access and refresh tokens within a transaction
func (s *Service) generateTokenPairTx(ctx context.Context, tx *sqlx.Tx, user *model.User) (*TokenPair, error) {
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

	// Hash refresh token for storage (use SHA-256, not bcrypt - JWT tokens exceed bcrypt's 72-byte limit)
	tokenHash := hash.HashToken(refreshToken)

	// Store refresh token within transaction
	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: refreshExpiresAt,
	}

	if err := s.tokenRepo.CreateWithCleanupTx(ctx, tx, rt, 5); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}
