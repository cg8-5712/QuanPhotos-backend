package category

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

// CategoryRepository handles category database operations
type CategoryRepository struct {
	*postgresql.BaseRepository
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// ListParams contains parameters for listing categories
type ListParams struct {
	Page     int
	PageSize int
}

// ListResult contains the result of listing categories
type ListResult struct {
	Categories []*model.Category
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// List retrieves all categories with photo counts
func (r *CategoryRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 50
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Count total
	var total int64
	err := r.DB().GetContext(ctx, &total, `SELECT COUNT(*) FROM categories`)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query categories with photo count
	query := `
		SELECT c.*, COALESCE(COUNT(p.id), 0) as photo_count
		FROM categories c
		LEFT JOIN photos p ON p.category_id = c.id AND p.status = 'approved'
		GROUP BY c.id
		ORDER BY c.sort_order ASC, c.id ASC
		LIMIT $1 OFFSET $2
	`

	var categories []*model.Category
	err = r.DB().SelectContext(ctx, &categories, query, params.PageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Categories: categories,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id int32) (*model.Category, error) {
	query := `
		SELECT c.*, COALESCE(COUNT(p.id), 0) as photo_count
		FROM categories c
		LEFT JOIN photos p ON p.category_id = c.id AND p.status = 'approved'
		WHERE c.id = $1
		GROUP BY c.id
	`
	var cat model.Category
	err := r.DB().GetContext(ctx, &cat, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, name, nameEN, description string, sortOrder int) (*model.Category, error) {
	query := `
		INSERT INTO categories (name, name_en, description, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, name_en, description, sort_order, created_at, updated_at
	`

	var cat model.Category
	err := r.DB().QueryRowContext(ctx, query, name, nameEN, toNullString(&description), sortOrder).Scan(
		&cat.ID, &cat.Name, &cat.NameEN, &cat.Description, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, postgresql.ErrDuplicateKey
		}
		return nil, err
	}
	return &cat, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, id int32, name, nameEN, description string, sortOrder int) (*model.Category, error) {
	query := `
		UPDATE categories
		SET name = $1, name_en = $2, description = $3, sort_order = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, name, name_en, description, sort_order, created_at, updated_at
	`

	var cat model.Category
	err := r.DB().QueryRowContext(ctx, query, name, nameEN, toNullString(&description), sortOrder, id).Scan(
		&cat.ID, &cat.Name, &cat.NameEN, &cat.Description, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, postgresql.ErrDuplicateKey
		}
		return nil, err
	}
	return &cat, nil
}

// Delete deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id int32) error {
	result, err := r.DB().ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
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

// Exists checks if a category exists
func (r *CategoryRepository) Exists(ctx context.Context, id int32) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`, id)
	return exists, err
}

// GetPhotoCount returns the number of approved photos in a category
func (r *CategoryRepository) GetPhotoCount(ctx context.Context, id int32) (int, error) {
	var count int
	err := r.DB().GetContext(ctx, &count, `SELECT COUNT(*) FROM photos WHERE category_id = $1 AND status = 'approved'`, id)
	return count, err
}

// toNullString helper function
func toNullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// ListPhotosParams contains parameters for listing photos in a category
type ListPhotosParams struct {
	CategoryID int32
	Page       int
	PageSize   int
	SortBy     string // created_at, like_count, view_count
	SortOrder  string // asc, desc
}

// ListPhotosResult contains the result of listing photos
type ListPhotosResult struct {
	Photos     []*model.Photo
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ListPhotos retrieves photos in a category
func (r *CategoryRepository) ListPhotos(ctx context.Context, params ListPhotosParams) (*ListPhotosResult, error) {
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
	var total int64
	err := r.DB().GetContext(ctx, &total, `SELECT COUNT(*) FROM photos WHERE category_id = $1 AND status = 'approved'`, params.CategoryID)
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
		SELECT * FROM photos
		WHERE category_id = $1 AND status = 'approved'
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, params.SortBy, params.SortOrder)

	var photos []*model.Photo
	err = r.DB().SelectContext(ctx, &photos, query, params.CategoryID, params.PageSize, offset)
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
