package auth

import (
	"context"
	"errors"

	"QuanPhotos/internal/pkg/hash"
	"QuanPhotos/internal/pkg/jwt"
	"QuanPhotos/internal/repository/postgresql"
)

var (
	ErrInvalidToken  = errors.New("invalid refresh token")
	ErrTokenExpired  = errors.New("refresh token has expired")
	ErrUserNotActive = errors.New("user is not active")
)

// Refresh refreshes the access token using a refresh token
func (s *Service) Refresh(ctx context.Context, req *RefreshRequest) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	// Check if user is still active
	if !user.IsActive() {
		return nil, ErrUserNotActive
	}

	// Verify token exists in database (check any valid token for user)
	tokens, err := s.tokenRepo.GetValidByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Check if the provided refresh token matches any stored token
	tokenValid := false
	var matchedTokenHash string
	for _, t := range tokens {
		if hash.CheckPassword(req.RefreshToken, t.TokenHash) {
			tokenValid = true
			matchedTokenHash = t.TokenHash
			break
		}
	}

	if !tokenValid {
		return nil, ErrInvalidToken
	}

	// Delete the old refresh token
	_ = s.tokenRepo.DeleteByTokenHash(ctx, matchedTokenHash)

	// Generate new token pair
	return s.generateTokenPair(ctx, user)
}
