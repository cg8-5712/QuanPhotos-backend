package user

import (
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// UserRepository handles user database operations
type UserRepository struct {
	*postgresql.BaseRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}
