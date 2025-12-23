package token

import (
	"context"

	"QuanPhotos/internal/repository/postgresql"
)

// Delete deletes a refresh token by ID
func (r *TokenRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`

	result, err := r.DB().ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrTokenNotFound
	}

	return nil
}

// DeleteByTokenHash deletes a refresh token by its hash
func (r *TokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	result, err := r.DB().ExecContext(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrTokenNotFound
	}

	return nil
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *TokenRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.DB().ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired deletes all expired refresh tokens
func (r *TokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`

	result, err := r.DB().ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
