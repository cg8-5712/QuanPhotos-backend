package token

import (
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// TokenRepository handles refresh token database operations
type TokenRepository struct {
	*postgresql.BaseRepository
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *sqlx.DB) *TokenRepository {
	return &TokenRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}
