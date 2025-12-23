package auth

import (
	"context"

	"QuanPhotos/internal/pkg/hash"
)

// Logout invalidates a refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	// Get all tokens for the user from the refresh token claims
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		// Even if token is invalid, we consider logout successful
		return nil
	}

	// Get user's tokens and find matching one
	tokens, err := s.tokenRepo.GetValidByUserID(ctx, claims.UserID)
	if err != nil {
		return nil
	}

	// Find and delete the matching token
	for _, t := range tokens {
		if hash.CheckPassword(refreshToken, t.TokenHash) {
			_ = s.tokenRepo.DeleteByTokenHash(ctx, t.TokenHash)
			break
		}
	}

	return nil
}

// LogoutAll invalidates all refresh tokens for a user
func (s *Service) LogoutAll(ctx context.Context, userID int64) error {
	return s.tokenRepo.DeleteByUserID(ctx, userID)
}
