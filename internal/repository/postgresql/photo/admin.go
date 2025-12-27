package photo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
)

// ReviewListParams contains parameters for listing pending reviews
type ReviewListParams struct {
	Page     int
	PageSize int
	Status   string // pending, ai_passed, ai_rejected
}

// ReviewListResult contains the result of listing pending reviews
type ReviewListResult struct {
	Photos     []*model.Photo
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ListPendingReviews retrieves photos pending manual review
func (r *PhotoRepository) ListPendingReviews(ctx context.Context, params ReviewListParams) (*ReviewListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Default: show ai_passed photos (waiting for manual review)
	if params.Status == "" {
		params.Status = string(model.PhotoStatusAIPassed)
	}

	if params.Status != "all" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	} else {
		// Show all pending/ai_passed/ai_rejected
		conditions = append(conditions, fmt.Sprintf("status IN ($%d, $%d, $%d)", argIndex, argIndex+1, argIndex+2))
		args = append(args, model.PhotoStatusPending, model.PhotoStatusAIPassed, model.PhotoStatusAIRejected)
		argIndex += 3
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM photos %s", whereClause)
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery, args...)
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
		%s
		ORDER BY created_at ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var photos []*model.Photo
	err = r.DB().SelectContext(ctx, &photos, query, args...)
	if err != nil {
		return nil, err
	}

	return &ReviewListResult{
		Photos:     photos,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ReviewPhoto performs a manual review on a photo
func (r *PhotoRepository) ReviewPhoto(ctx context.Context, photoID, reviewerID int64, action, reason string) error {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Determine new status
	var newStatus model.PhotoStatus
	if action == "approve" {
		newStatus = model.PhotoStatusApproved
	} else {
		newStatus = model.PhotoStatusRejected
	}

	// Update photo status
	updateQuery := `UPDATE photos SET status = $1, updated_at = NOW() WHERE id = $2`
	if newStatus == model.PhotoStatusApproved {
		updateQuery = `UPDATE photos SET status = $1, approved_at = NOW(), updated_at = NOW() WHERE id = $2`
	}

	result, err := tx.ExecContext(ctx, updateQuery, newStatus, photoID)
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

	// Insert review record
	insertQuery := `
		INSERT INTO photo_reviews (photo_id, reviewer_id, review_type, action, reason)
		VALUES ($1, $2, 'manual', $3, $4)
	`
	_, err = tx.ExecContext(ctx, insertQuery, photoID, reviewerID, action, toNullString(&reason))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AdminDeletePhoto deletes a photo with reason (admin action)
func (r *PhotoRepository) AdminDeletePhoto(ctx context.Context, photoID, adminID int64, reason string) error {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if photo exists
	var exists bool
	err = tx.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM photos WHERE id = $1)`, photoID)
	if err != nil {
		return err
	}
	if !exists {
		return postgresql.ErrNotFound
	}

	// Insert review record as rejection before deletion
	insertQuery := `
		INSERT INTO photo_reviews (photo_id, reviewer_id, review_type, action, reason)
		VALUES ($1, $2, 'manual', 'reject', $3)
	`
	_, err = tx.ExecContext(ctx, insertQuery, photoID, adminID, reason)
	if err != nil {
		return err
	}

	// Delete photo
	_, err = tx.ExecContext(ctx, `DELETE FROM photos WHERE id = $1`, photoID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// FeaturedPhoto represents a featured photo record
type FeaturedPhoto struct {
	ID         int64          `db:"id"`
	PhotoID    int64          `db:"photo_id"`
	AdminID    int64          `db:"admin_id"`
	Reason     sql.NullString `db:"reason"`
	SortOrder  int            `db:"sort_order"`
	FeaturedAt time.Time      `db:"featured_at"`
	ExpiresAt  sql.NullTime   `db:"expires_at"`
}

// AddFeatured adds a photo to featured list
func (r *PhotoRepository) AddFeatured(ctx context.Context, photoID, adminID int64, reason string, sortOrder int) error {
	query := `
		INSERT INTO featured_photos (photo_id, admin_id, reason, sort_order)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.DB().ExecContext(ctx, query, photoID, adminID, toNullString(&reason), sortOrder)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return postgresql.ErrDuplicateKey
		}
		return err
	}
	return nil
}

// RemoveFeatured removes a photo from featured list
func (r *PhotoRepository) RemoveFeatured(ctx context.Context, featuredID int64) error {
	result, err := r.DB().ExecContext(ctx, `DELETE FROM featured_photos WHERE id = $1`, featuredID)
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

// RemoveFeaturedByPhotoID removes a photo from featured list by photo ID
func (r *PhotoRepository) RemoveFeaturedByPhotoID(ctx context.Context, photoID int64) error {
	result, err := r.DB().ExecContext(ctx, `DELETE FROM featured_photos WHERE photo_id = $1`, photoID)
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

// IsFeaturedPhoto checks if a photo is featured
func (r *PhotoRepository) IsFeaturedPhoto(ctx context.Context, photoID int64) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM featured_photos WHERE photo_id = $1)`, photoID)
	return exists, err
}

// GetFeaturedByPhotoID gets featured record by photo ID
func (r *PhotoRepository) GetFeaturedByPhotoID(ctx context.Context, photoID int64) (*FeaturedPhoto, error) {
	var featured FeaturedPhoto
	err := r.DB().GetContext(ctx, &featured, `SELECT * FROM featured_photos WHERE photo_id = $1`, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &featured, nil
}

// Announcement represents an announcement record
type Announcement struct {
	ID          int64          `db:"id" json:"id"`
	AuthorID    int64          `db:"author_id" json:"author_id"`
	Title       string         `db:"title" json:"title"`
	Summary     sql.NullString `db:"summary" json:"-"`
	Content     string         `db:"content" json:"content"`
	Status      string         `db:"status" json:"status"`
	IsPinned    bool           `db:"is_pinned" json:"is_pinned"`
	PublishedAt sql.NullTime   `db:"published_at" json:"-"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

// CreateAnnouncement creates a new announcement
func (r *PhotoRepository) CreateAnnouncement(ctx context.Context, authorID int64, title, summary, content, status string, isPinned bool) (int64, error) {
	var publishedAt sql.NullTime
	if status == "published" {
		publishedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	query := `
		INSERT INTO announcements (author_id, title, summary, content, status, is_pinned, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int64
	err := r.DB().QueryRowContext(ctx, query, authorID, title, toNullString(&summary), content, status, isPinned, publishedAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetAnnouncementByID gets an announcement by ID
func (r *PhotoRepository) GetAnnouncementByID(ctx context.Context, id int64) (*Announcement, error) {
	var ann Announcement
	err := r.DB().GetContext(ctx, &ann, `SELECT * FROM announcements WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &ann, nil
}

// UpdateAnnouncement updates an announcement
func (r *PhotoRepository) UpdateAnnouncement(ctx context.Context, id int64, title, summary, content, status string, isPinned bool) error {
	var publishedAt sql.NullTime
	if status == "published" {
		// Check if already published
		var existingPublishedAt sql.NullTime
		err := r.DB().GetContext(ctx, &existingPublishedAt, `SELECT published_at FROM announcements WHERE id = $1`, id)
		if err != nil {
			return err
		}
		if existingPublishedAt.Valid {
			publishedAt = existingPublishedAt
		} else {
			publishedAt = sql.NullTime{Time: time.Now(), Valid: true}
		}
	}

	query := `
		UPDATE announcements
		SET title = $1, summary = $2, content = $3, status = $4, is_pinned = $5, published_at = $6, updated_at = NOW()
		WHERE id = $7
	`
	result, err := r.DB().ExecContext(ctx, query, title, toNullString(&summary), content, status, isPinned, publishedAt, id)
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

// DeleteAnnouncement deletes an announcement
func (r *PhotoRepository) DeleteAnnouncement(ctx context.Context, id int64) error {
	result, err := r.DB().ExecContext(ctx, `DELETE FROM announcements WHERE id = $1`, id)
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

// AnnouncementListParams contains parameters for listing announcements
type AnnouncementListParams struct {
	Page     int
	PageSize int
	Status   string // draft, published, all
}

// AnnouncementListResult contains the result of listing announcements
type AnnouncementListResult struct {
	Announcements []*Announcement
	Total         int64
	Page          int
	PageSize      int
	TotalPages    int
}

// ListAnnouncements lists announcements with pagination
func (r *PhotoRepository) ListAnnouncements(ctx context.Context, params AnnouncementListParams) (*AnnouncementListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Build WHERE clause
	whereClause := ""
	var args []interface{}
	argIndex := 1

	if params.Status != "" && params.Status != "all" {
		whereClause = fmt.Sprintf("WHERE status = $%d", argIndex)
		args = append(args, params.Status)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM announcements %s", whereClause)
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query announcements
	query := fmt.Sprintf(`
		SELECT * FROM announcements
		%s
		ORDER BY is_pinned DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var announcements []*Announcement
	err = r.DB().SelectContext(ctx, &announcements, query, args...)
	if err != nil {
		return nil, err
	}

	return &AnnouncementListResult{
		Announcements: announcements,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    totalPages,
	}, nil
}

// ============================================
// Public Methods
// ============================================

// FeaturedListParams contains parameters for listing featured photos
type FeaturedListParams struct {
	Page     int
	PageSize int
}

// FeaturedListResult contains the result of listing featured photos
type FeaturedListResult struct {
	Photos     []*model.Photo
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ListFeaturedPhotos lists featured photos publicly
func (r *PhotoRepository) ListFeaturedPhotos(ctx context.Context, params FeaturedListParams) (*FeaturedListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Count total
	countQuery := `
		SELECT COUNT(*) FROM featured_photos fp
		INNER JOIN photos p ON fp.photo_id = p.id
		WHERE p.status = 'approved' AND (fp.expires_at IS NULL OR fp.expires_at > NOW())
	`
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query featured photos
	query := `
		SELECT p.* FROM photos p
		INNER JOIN featured_photos fp ON p.id = fp.photo_id
		WHERE p.status = 'approved' AND (fp.expires_at IS NULL OR fp.expires_at > NOW())
		ORDER BY fp.sort_order ASC, fp.featured_at DESC
		LIMIT $1 OFFSET $2
	`

	var photos []*model.Photo
	err = r.DB().SelectContext(ctx, &photos, query, params.PageSize, offset)
	if err != nil {
		return nil, err
	}

	return &FeaturedListResult{
		Photos:     photos,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListPublicAnnouncements lists published announcements
func (r *PhotoRepository) ListPublicAnnouncements(ctx context.Context, params AnnouncementListParams) (*AnnouncementListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Count total (only published)
	countQuery := `SELECT COUNT(*) FROM announcements WHERE status = 'published'`
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query announcements
	query := `
		SELECT * FROM announcements
		WHERE status = 'published'
		ORDER BY is_pinned DESC, published_at DESC
		LIMIT $1 OFFSET $2
	`

	var announcements []*Announcement
	err = r.DB().SelectContext(ctx, &announcements, query, params.PageSize, offset)
	if err != nil {
		return nil, err
	}

	return &AnnouncementListResult{
		Announcements: announcements,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    totalPages,
	}, nil
}

// GetPublicAnnouncementByID gets a published announcement by ID
func (r *PhotoRepository) GetPublicAnnouncementByID(ctx context.Context, id int64) (*Announcement, error) {
	var ann Announcement
	err := r.DB().GetContext(ctx, &ann, `SELECT * FROM announcements WHERE id = $1 AND status = 'published'`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &ann, nil
}
