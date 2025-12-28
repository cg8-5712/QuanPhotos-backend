package tag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// TagRepository handles tag database operations
type TagRepository struct {
	*postgresql.BaseRepository
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *sqlx.DB) *TagRepository {
	return &TagRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// ListParams contains parameters for listing tags
type ListParams struct {
	Page     int
	PageSize int
	OrderBy  string // photo_count, name, created_at
}

// ListResult contains the result of listing tags
type ListResult struct {
	Tags       []*model.Tag
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// List retrieves tags ordered by photo count (popular tags)
func (r *TagRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 50
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Validate order by
	validOrderBy := map[string]bool{"photo_count": true, "name": true, "created_at": true}
	if !validOrderBy[params.OrderBy] {
		params.OrderBy = "photo_count"
	}

	orderClause := params.OrderBy
	if params.OrderBy == "photo_count" {
		orderClause = "photo_count DESC"
	} else if params.OrderBy == "name" {
		orderClause = "name ASC"
	} else {
		orderClause = "created_at DESC"
	}

	// Count total
	var total int64
	err := r.DB().GetContext(ctx, &total, `SELECT COUNT(*) FROM tags`)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query tags
	query := fmt.Sprintf(`
		SELECT * FROM tags
		ORDER BY %s
		LIMIT $1 OFFSET $2
	`, orderClause)

	var tags []*model.Tag
	err = r.DB().SelectContext(ctx, &tags, query, params.PageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Tags:       tags,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Search searches tags by keyword
func (r *TagRepository) Search(ctx context.Context, keyword string, limit int) ([]*model.Tag, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	query := `
		SELECT * FROM tags
		WHERE name ILIKE $1
		ORDER BY photo_count DESC, name ASC
		LIMIT $2
	`

	var tags []*model.Tag
	err := r.DB().SelectContext(ctx, &tags, query, "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// GetByID retrieves a tag by ID
func (r *TagRepository) GetByID(ctx context.Context, id int32) (*model.Tag, error) {
	var tag model.Tag
	err := r.DB().GetContext(ctx, &tag, `SELECT * FROM tags WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &tag, nil
}

// GetByName retrieves a tag by name
func (r *TagRepository) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.DB().GetContext(ctx, &tag, `SELECT * FROM tags WHERE name = $1`, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &tag, nil
}

// Create creates a new tag
func (r *TagRepository) Create(ctx context.Context, name string) (*model.Tag, error) {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		RETURNING id, name, photo_count, created_at
	`

	var tag model.Tag
	err := r.DB().QueryRowContext(ctx, query, name).Scan(&tag.ID, &tag.Name, &tag.PhotoCount, &tag.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, postgresql.ErrDuplicateKey
		}
		return nil, err
	}
	return &tag, nil
}

// GetOrCreate gets a tag by name or creates it if not exists
func (r *TagRepository) GetOrCreate(ctx context.Context, name string) (*model.Tag, error) {
	// Try to get first
	tag, err := r.GetByName(ctx, name)
	if err == nil {
		return tag, nil
	}
	if !errors.Is(err, postgresql.ErrNotFound) {
		return nil, err
	}

	// Create if not exists
	return r.Create(ctx, name)
}

// Delete deletes a tag
func (r *TagRepository) Delete(ctx context.Context, id int32) error {
	result, err := r.DB().ExecContext(ctx, `DELETE FROM tags WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return postgresql.ErrNotFound
	}
	return nil
}

// Exists checks if a tag exists
func (r *TagRepository) Exists(ctx context.Context, id int32) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)`, id)
	return exists, err
}

// ListPhotosParams contains parameters for listing photos with a tag
type ListPhotosParams struct {
	TagID     int32
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// ListPhotosResult contains the result of listing photos
type ListPhotosResult struct {
	Photos     []*model.Photo
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ListPhotos retrieves photos with a specific tag
func (r *TagRepository) ListPhotos(ctx context.Context, params ListPhotosParams) (*ListPhotosResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Validate sort parameters
	validSortBy := map[string]bool{"created_at": true, "like_count": true, "view_count": true}
	if !validSortBy[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "desc"
	}

	// Count total
	countQuery := `
		SELECT COUNT(*) FROM photos p
		JOIN photo_tags pt ON pt.photo_id = p.id
		WHERE pt.tag_id = $1 AND p.status = 'approved'
	`
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery, params.TagID)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query photos
	query := fmt.Sprintf(`
		SELECT p.* FROM photos p
		JOIN photo_tags pt ON pt.photo_id = p.id
		WHERE pt.tag_id = $1 AND p.status = 'approved'
		ORDER BY p.%s %s
		LIMIT $2 OFFSET $3
	`, params.SortBy, params.SortOrder)

	var photos []*model.Photo
	err = r.DB().SelectContext(ctx, &photos, query, params.TagID, params.PageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ListPhotosResult{
		Photos:     photos,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPhotoTags retrieves all tags for a photo
func (r *TagRepository) GetPhotoTags(ctx context.Context, photoID int64) ([]*model.Tag, error) {
	query := `
		SELECT t.* FROM tags t
		JOIN photo_tags pt ON pt.tag_id = t.id
		WHERE pt.photo_id = $1
		ORDER BY t.name ASC
	`

	var tags []*model.Tag
	err := r.DB().SelectContext(ctx, &tags, query, photoID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// AddPhotoTag adds a tag to a photo
func (r *TagRepository) AddPhotoTag(ctx context.Context, photoID int64, tagID int32) error {
	_, err := r.DB().ExecContext(ctx, `INSERT INTO photo_tags (photo_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, photoID, tagID)
	return err
}

// RemovePhotoTag removes a tag from a photo
func (r *TagRepository) RemovePhotoTag(ctx context.Context, photoID int64, tagID int32) error {
	_, err := r.DB().ExecContext(ctx, `DELETE FROM photo_tags WHERE photo_id = $1 AND tag_id = $2`, photoID, tagID)
	return err
}

// SetPhotoTags replaces all tags for a photo
func (r *TagRepository) SetPhotoTags(ctx context.Context, photoID int64, tagIDs []int32) error {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing tags
	_, err = tx.ExecContext(ctx, `DELETE FROM photo_tags WHERE photo_id = $1`, photoID)
	if err != nil {
		return err
	}

	// Insert new tags
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			_, err = tx.ExecContext(ctx, `INSERT INTO photo_tags (photo_id, tag_id) VALUES ($1, $2)`, photoID, tagID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetUserMap retrieves a map of users by their IDs
func (r *TagRepository) GetUserMap(ctx context.Context, userIDs []int64) (map[int64]*model.User, error) {
	if len(userIDs) == 0 {
		return make(map[int64]*model.User), nil
	}

	query, args, err := sqlx.In(`SELECT * FROM users WHERE id IN (?)`, userIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB().Rebind(query)

	var users []*model.User
	err = r.DB().SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*model.User)
	for _, u := range users {
		result[u.ID] = u
	}

	return result, nil
}
