package notification

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	TypeLike     NotificationType = "like"
	TypeComment  NotificationType = "comment"
	TypeReply    NotificationType = "reply"
	TypeFollow   NotificationType = "follow"
	TypeShare    NotificationType = "share"
	TypeFeatured NotificationType = "featured"
	TypeReview   NotificationType = "review"
	TypeSystem   NotificationType = "system"
	TypeMessage  NotificationType = "message"
)

// Notification represents a notification record
type Notification struct {
	ID               int64          `db:"id" json:"id"`
	UserID           int64          `db:"user_id" json:"user_id"`
	ActorID          sql.NullInt64  `db:"actor_id" json:"-"`
	Type             string         `db:"type" json:"type"`
	Title            string         `db:"title" json:"title"`
	Content          sql.NullString `db:"content" json:"-"`
	RelatedPhotoID   sql.NullInt64  `db:"related_photo_id" json:"-"`
	RelatedCommentID sql.NullInt64  `db:"related_comment_id" json:"-"`
	IsRead           bool           `db:"is_read" json:"is_read"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
}

// NotificationRepository handles notification database operations
type NotificationRepository struct {
	*postgresql.BaseRepository
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// ListParams contains parameters for listing notifications
type ListParams struct {
	UserID   int64
	Type     string // filter by type
	IsRead   *bool  // filter by read status
	Page     int
	PageSize int
}

// ListResult contains the result of listing notifications
type ListResult struct {
	Notifications []*Notification
	Total         int64
	Page          int
	PageSize      int
	TotalPages    int
}

// List retrieves notifications for a user
func (r *NotificationRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
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

	conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
	args = append(args, params.UserID)
	argIndex++

	if params.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, params.Type)
		argIndex++
	}

	if params.IsRead != nil {
		conditions = append(conditions, fmt.Sprintf("is_read = $%d", argIndex))
		args = append(args, *params.IsRead)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", whereClause)
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

	// Query notifications
	query := fmt.Sprintf(`
		SELECT * FROM notifications
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var notifications []*Notification
	err = r.DB().SelectContext(ctx, &notifications, query, args...)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Notifications: notifications,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    totalPages,
	}, nil
}

// GetUnreadCount returns the unread notification count for a user
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.DB().GetContext(ctx, &count, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE
	`, userID)
	return count, err
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID int64) error {
	result, err := r.DB().ExecContext(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2
	`, id, userID)
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

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	_, err := r.DB().ExecContext(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE
	`, userID)
	return err
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, n *Notification) (*Notification, error) {
	query := `
		INSERT INTO notifications (user_id, actor_id, type, title, content, related_photo_id, related_comment_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`
	var notification Notification
	err := r.DB().GetContext(ctx, &notification, query,
		n.UserID, n.ActorID, n.Type, n.Title, n.Content, n.RelatedPhotoID, n.RelatedCommentID)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// CreateBatch creates multiple notifications
func (r *NotificationRepository) CreateBatch(ctx context.Context, notifications []*Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := `
		INSERT INTO notifications (user_id, actor_id, type, title, content, related_photo_id, related_comment_id)
		VALUES (:user_id, :actor_id, :type, :title, :content, :related_photo_id, :related_comment_id)
	`
	_, err := r.DB().NamedExecContext(ctx, query, notifications)
	return err
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id, userID int64) error {
	result, err := r.DB().ExecContext(ctx, `
		DELETE FROM notifications WHERE id = $1 AND user_id = $2
	`, id, userID)
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

// GetUserMap retrieves a map of users by their IDs
func (r *NotificationRepository) GetUserMap(ctx context.Context, userIDs []int64) (map[int64]*model.User, error) {
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

// Exists checks if a notification exists
func (r *NotificationRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1)`, id)
	return exists, err
}
