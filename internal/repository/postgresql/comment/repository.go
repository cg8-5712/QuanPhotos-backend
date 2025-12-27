package comment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// Comment represents a photo comment
type Comment struct {
	ID         int64          `db:"id" json:"id"`
	PhotoID    int64          `db:"photo_id" json:"photo_id"`
	UserID     int64          `db:"user_id" json:"user_id"`
	ParentID   sql.NullInt64  `db:"parent_id" json:"-"`
	Content    string         `db:"content" json:"content"`
	LikeCount  int            `db:"like_count" json:"like_count"`
	ReplyCount int            `db:"reply_count" json:"reply_count"`
	Status     string         `db:"status" json:"status"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}

// CommentRepository handles comment database operations
type CommentRepository struct {
	*postgresql.BaseRepository
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *sqlx.DB) *CommentRepository {
	return &CommentRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// ListParams contains parameters for listing comments
type ListParams struct {
	PhotoID  int64
	ParentID *int64 // nil = top-level comments only
	Page     int
	PageSize int
	SortBy   string // created_at, like_count
}

// ListResult contains the result of listing comments
type ListResult struct {
	Comments   []*Comment
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// List retrieves comments for a photo
func (r *CommentRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
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

	conditions = append(conditions, fmt.Sprintf("photo_id = $%d", argIndex))
	args = append(args, params.PhotoID)
	argIndex++

	conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
	args = append(args, "visible")
	argIndex++

	if params.ParentID == nil {
		conditions = append(conditions, "parent_id IS NULL")
	} else {
		conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
		args = append(args, *params.ParentID)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM photo_comments %s", whereClause)
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

	// Determine sort order
	sortClause := "created_at DESC"
	if params.SortBy == "like_count" {
		sortClause = "like_count DESC, created_at DESC"
	}

	// Query comments
	query := fmt.Sprintf(`
		SELECT * FROM photo_comments
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var comments []*Comment
	err = r.DB().SelectContext(ctx, &comments, query, args...)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Comments:   comments,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(ctx context.Context, id int64) (*Comment, error) {
	var comment Comment
	err := r.DB().GetContext(ctx, &comment, `SELECT * FROM photo_comments WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &comment, nil
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, photoID, userID int64, parentID *int64, content string) (*Comment, error) {
	var comment Comment
	query := `
		INSERT INTO photo_comments (photo_id, user_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`

	var parentIDValue sql.NullInt64
	if parentID != nil {
		parentIDValue = sql.NullInt64{Int64: *parentID, Valid: true}
	}

	err := r.DB().GetContext(ctx, &comment, query, photoID, userID, parentIDValue, content)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// Delete soft deletes a comment
func (r *CommentRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.DB().ExecContext(ctx, `UPDATE photo_comments SET status = 'deleted', updated_at = NOW() WHERE id = $1`, id)
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

// HardDelete permanently deletes a comment
func (r *CommentRepository) HardDelete(ctx context.Context, id int64) error {
	result, err := r.DB().ExecContext(ctx, `DELETE FROM photo_comments WHERE id = $1`, id)
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

// Exists checks if a comment exists
func (r *CommentRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM photo_comments WHERE id = $1 AND status = 'visible')`, id)
	return exists, err
}

// IsOwnedBy checks if a comment is owned by a user
func (r *CommentRepository) IsOwnedBy(ctx context.Context, commentID, userID int64) (bool, error) {
	var isOwner bool
	err := r.DB().GetContext(ctx, &isOwner, `SELECT EXISTS(SELECT 1 FROM photo_comments WHERE id = $1 AND user_id = $2)`, commentID, userID)
	return isOwner, err
}

// AddLike adds a like to a comment
func (r *CommentRepository) AddLike(ctx context.Context, userID, commentID int64) error {
	_, err := r.DB().ExecContext(ctx, `INSERT INTO comment_likes (user_id, comment_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, commentID)
	return err
}

// RemoveLike removes a like from a comment
func (r *CommentRepository) RemoveLike(ctx context.Context, userID, commentID int64) error {
	result, err := r.DB().ExecContext(ctx, `DELETE FROM comment_likes WHERE user_id = $1 AND comment_id = $2`, userID, commentID)
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

// IsLiked checks if a user has liked a comment
func (r *CommentRepository) IsLiked(ctx context.Context, userID, commentID int64) (bool, error) {
	var isLiked bool
	err := r.DB().GetContext(ctx, &isLiked, `SELECT EXISTS(SELECT 1 FROM comment_likes WHERE user_id = $1 AND comment_id = $2)`, userID, commentID)
	return isLiked, err
}

// GetUserMap retrieves a map of users by their IDs
func (r *CommentRepository) GetUserMap(ctx context.Context, userIDs []int64) (map[int64]*model.User, error) {
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

// GetReplies retrieves replies for a comment
func (r *CommentRepository) GetReplies(ctx context.Context, parentID int64, limit int) ([]*Comment, error) {
	if limit < 1 {
		limit = 3
	}
	if limit > 20 {
		limit = 20
	}

	var comments []*Comment
	err := r.DB().SelectContext(ctx, &comments, `
		SELECT * FROM photo_comments
		WHERE parent_id = $1 AND status = 'visible'
		ORDER BY created_at ASC
		LIMIT $2
	`, parentID, limit)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// GetPhotoUserID gets the user ID of the photo owner
func (r *CommentRepository) GetPhotoUserID(ctx context.Context, photoID int64) (int64, error) {
	var userID int64
	err := r.DB().GetContext(ctx, &userID, `SELECT user_id FROM photos WHERE id = $1`, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, postgresql.ErrNotFound
		}
		return 0, err
	}
	return userID, nil
}

// PhotoExists checks if a photo exists
func (r *CommentRepository) PhotoExists(ctx context.Context, photoID int64) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM photos WHERE id = $1)`, photoID)
	return exists, err
}
