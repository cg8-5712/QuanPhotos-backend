package photo

import (
	"context"
	"database/sql"
	"errors"

	"QuanPhotos/internal/repository/postgresql"
)

// AddFavorite adds a photo to user's favorites
func (r *PhotoRepository) AddFavorite(ctx context.Context, userID, photoID int64) error {
	query := `INSERT INTO favorites (user_id, photo_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.DB().ExecContext(ctx, query, userID, photoID)
	return err
}

// RemoveFavorite removes a photo from user's favorites
func (r *PhotoRepository) RemoveFavorite(ctx context.Context, userID, photoID int64) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND photo_id = $2`
	result, err := r.DB().ExecContext(ctx, query, userID, photoID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrNotFound
	}

	return nil
}

// AddLike adds a like to a photo
func (r *PhotoRepository) AddLike(ctx context.Context, userID, photoID int64) error {
	query := `INSERT INTO photo_likes (user_id, photo_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.DB().ExecContext(ctx, query, userID, photoID)
	return err
}

// RemoveLike removes a like from a photo
func (r *PhotoRepository) RemoveLike(ctx context.Context, userID, photoID int64) error {
	query := `DELETE FROM photo_likes WHERE user_id = $1 AND photo_id = $2`
	result, err := r.DB().ExecContext(ctx, query, userID, photoID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrNotFound
	}

	return nil
}

// Exists checks if a photo exists
func (r *PhotoRepository) Exists(ctx context.Context, photoID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM photos WHERE id = $1)`
	err := r.DB().GetContext(ctx, &exists, query, photoID)
	return exists, err
}

// IsOwnedBy checks if a photo is owned by a user
func (r *PhotoRepository) IsOwnedBy(ctx context.Context, photoID, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM photos WHERE id = $1 AND user_id = $2)`
	err := r.DB().GetContext(ctx, &exists, query, photoID, userID)
	return exists, err
}

// Delete soft deletes a photo (or hard delete based on requirements)
func (r *PhotoRepository) Delete(ctx context.Context, photoID int64) error {
	query := `DELETE FROM photos WHERE id = $1`
	result, err := r.DB().ExecContext(ctx, query, photoID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrNotFound
	}

	return nil
}

// GetFilePaths retrieves file paths for a photo (for deletion)
func (r *PhotoRepository) GetFilePaths(ctx context.Context, photoID int64) (filePath, thumbnailPath, rawFilePath string, err error) {
	var photo struct {
		FilePath      string         `db:"file_path"`
		ThumbnailPath sql.NullString `db:"thumbnail_path"`
		RawFilePath   sql.NullString `db:"raw_file_path"`
	}

	query := `SELECT file_path, thumbnail_path, raw_file_path FROM photos WHERE id = $1`
	err = r.DB().GetContext(ctx, &photo, query, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", postgresql.ErrNotFound
		}
		return "", "", "", err
	}

	filePath = photo.FilePath
	if photo.ThumbnailPath.Valid {
		thumbnailPath = photo.ThumbnailPath.String
	}
	if photo.RawFilePath.Valid {
		rawFilePath = photo.RawFilePath.String
	}

	return filePath, thumbnailPath, rawFilePath, nil
}
