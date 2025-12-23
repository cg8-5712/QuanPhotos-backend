package postgresql

import (
	"github.com/jmoiron/sqlx"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// DB returns the database connection
func (r *BaseRepository) DB() *sqlx.DB {
	return r.db
}
