package token

import (
	"context"

	"QuanPhotos/internal/model"

	"github.com/jmoiron/sqlx"
)

// Create creates a new refresh token
func (r *TokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	return r.DB().QueryRowxContext(ctx, query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

// CreateTx creates a new refresh token within a transaction
func (r *TokenRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	return tx.QueryRowxContext(ctx, query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

// CreateWithCleanup creates a new token and removes old tokens for the user
func (r *TokenRepository) CreateWithCleanup(ctx context.Context, token *model.RefreshToken, maxTokensPerUser int) error {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.createWithCleanupTx(ctx, tx, token, maxTokensPerUser); err != nil {
		return err
	}

	return tx.Commit()
}

// CreateWithCleanupTx creates a new token and removes old tokens within an existing transaction
func (r *TokenRepository) CreateWithCleanupTx(ctx context.Context, tx *sqlx.Tx, token *model.RefreshToken, maxTokensPerUser int) error {
	return r.createWithCleanupTx(ctx, tx, token, maxTokensPerUser)
}

// createWithCleanupTx is the internal implementation for creating token with cleanup
func (r *TokenRepository) createWithCleanupTx(ctx context.Context, tx *sqlx.Tx, token *model.RefreshToken, maxTokensPerUser int) error {
	// Delete expired tokens for user
	_, err := tx.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1 AND expires_at < NOW()`,
		token.UserID,
	)
	if err != nil {
		return err
	}

	// If max tokens exceeded, delete oldest ones
	if maxTokensPerUser > 0 {
		_, err = tx.ExecContext(ctx, `
			DELETE FROM refresh_tokens
			WHERE id IN (
				SELECT id FROM refresh_tokens
				WHERE user_id = $1
				ORDER BY created_at DESC
				OFFSET $2
			)
		`, token.UserID, maxTokensPerUser-1)
		if err != nil {
			return err
		}
	}

	// Create new token
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return tx.QueryRowxContext(ctx, query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}
