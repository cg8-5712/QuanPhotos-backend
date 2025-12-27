package share

import (
	"context"
	"database/sql"
	"time"

	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// Share represents a photo share record
type Share struct {
	ID        int64          `db:"id" json:"id"`
	PhotoID   int64          `db:"photo_id" json:"photo_id"`
	UserID    int64          `db:"user_id" json:"user_id"`
	Content   sql.NullString `db:"content" json:"-"`
	ShareType string         `db:"share_type" json:"share_type"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}

// ShareRepository handles share database operations
type ShareRepository struct {
	*postgresql.BaseRepository
}

// NewShareRepository creates a new share repository
func NewShareRepository(db *sqlx.DB) *ShareRepository {
	return &ShareRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// Create creates a new share record
func (r *ShareRepository) Create(ctx context.Context, photoID, userID int64, shareType, content string) (*Share, error) {
	var share Share
	query := `
		INSERT INTO photo_shares (photo_id, user_id, share_type, content)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`

	var contentValue sql.NullString
	if content != "" {
		contentValue = sql.NullString{String: content, Valid: true}
	}

	// Map platform to share_type
	actualShareType := "external"
	if shareType == "link" {
		actualShareType = "internal"
	}

	err := r.DB().GetContext(ctx, &share, query, photoID, userID, actualShareType, contentValue)
	if err != nil {
		return nil, err
	}

	return &share, nil
}

// GetShareCount returns the total share count for a photo
func (r *ShareRepository) GetShareCount(ctx context.Context, photoID int64) (int, error) {
	var count int
	err := r.DB().GetContext(ctx, &count, `SELECT COUNT(*) FROM photo_shares WHERE photo_id = $1`, photoID)
	return count, err
}

// PhotoExists checks if a photo exists and is visible
func (r *ShareRepository) PhotoExists(ctx context.Context, photoID int64) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM photos WHERE id = $1 AND status = 'approved')`, photoID)
	return exists, err
}

// IncrementShareCount increments the share count for a photo
// Note: This is handled by database trigger, so this method is deprecated
func (r *ShareRepository) IncrementShareCount(ctx context.Context, photoID int64) error {
	// The database has a trigger that automatically increments share_count
	// So we don't need to do anything here
	return nil
}
