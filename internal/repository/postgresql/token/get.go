package token

import (
	"context"
	"database/sql"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
)

// GetByTokenHash retrieves a refresh token by its hash
func (r *TokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE token_hash = $1`

	err := r.DB().GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrTokenNotFound
		}
		return nil, err
	}

	return &token, nil
}

// GetByUserID retrieves all refresh tokens for a user
func (r *TokenRepository) GetByUserID(ctx context.Context, userID int64) ([]*model.RefreshToken, error) {
	var tokens []*model.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.DB().SelectContext(ctx, &tokens, query, userID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// GetValidByUserID retrieves all valid (non-expired) refresh tokens for a user
func (r *TokenRepository) GetValidByUserID(ctx context.Context, userID int64) ([]*model.RefreshToken, error) {
	var tokens []*model.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE user_id = $1 AND expires_at > NOW() ORDER BY created_at DESC`

	err := r.DB().SelectContext(ctx, &tokens, query, userID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
