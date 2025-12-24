package photo

import (
	"context"
	"database/sql"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
)

// GetByID retrieves a photo by ID
func (r *PhotoRepository) GetByID(ctx context.Context, id int64) (*model.Photo, error) {
	var photo model.Photo
	query := `SELECT * FROM photos WHERE id = $1`

	err := r.DB().GetContext(ctx, &photo, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}

	return &photo, nil
}

// GetUserByPhotoID retrieves user info for a photo
func (r *PhotoRepository) GetUserByPhotoID(ctx context.Context, photoID int64) (*model.User, error) {
	var user model.User
	query := `
		SELECT u.* FROM users u
		INNER JOIN photos p ON p.user_id = u.id
		WHERE p.id = $1
	`

	err := r.DB().GetContext(ctx, &user, query, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetCategoryByID retrieves a category by ID
func (r *PhotoRepository) GetCategoryByID(ctx context.Context, id int32) (*model.Category, error) {
	var category model.Category
	query := `SELECT * FROM categories WHERE id = $1`

	err := r.DB().GetContext(ctx, &category, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}

	return &category, nil
}

// GetTagsByPhotoID retrieves tags for a photo
func (r *PhotoRepository) GetTagsByPhotoID(ctx context.Context, photoID int64) ([]string, error) {
	var tags []string
	query := `
		SELECT t.name FROM tags t
		INNER JOIN photo_tags pt ON pt.tag_id = t.id
		WHERE pt.photo_id = $1
		ORDER BY t.name
	`

	err := r.DB().SelectContext(ctx, &tags, query, photoID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// IsFavorited checks if user has favorited a photo
func (r *PhotoRepository) IsFavorited(ctx context.Context, userID, photoID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = $1 AND photo_id = $2`

	err := r.DB().GetContext(ctx, &count, query, userID, photoID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsLiked checks if user has liked a photo
func (r *PhotoRepository) IsLiked(ctx context.Context, userID, photoID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM photo_likes WHERE user_id = $1 AND photo_id = $2`

	err := r.DB().GetContext(ctx, &count, query, userID, photoID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IncrementViewCount increments photo view count
func (r *PhotoRepository) IncrementViewCount(ctx context.Context, photoID int64) error {
	query := `UPDATE photos SET view_count = view_count + 1 WHERE id = $1`
	_, err := r.DB().ExecContext(ctx, query, photoID)
	return err
}
