package photo

import (
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// PhotoRepository handles photo database operations
type PhotoRepository struct {
	*postgresql.BaseRepository
}

// NewPhotoRepository creates a new photo repository
func NewPhotoRepository(db *sqlx.DB) *PhotoRepository {
	return &PhotoRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}
