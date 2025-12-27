package notification

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/notification"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

// Service handles notification business logic
type Service struct {
	notifRepo *notification.NotificationRepository
}

// New creates a new notification service
func New(notifRepo *notification.NotificationRepository) *Service {
	return &Service{
		notifRepo: notifRepo,
	}
}

// UserBrief represents brief user info
type UserBrief struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

// NotificationItem represents a notification in response
type NotificationItem struct {
	ID               int64       `json:"id"`
	Type             string      `json:"type"`
	Title            string      `json:"title"`
	Content          *string     `json:"content,omitempty"`
	Actor            *UserBrief  `json:"actor,omitempty"`
	RelatedPhotoID   *int64      `json:"related_photo_id,omitempty"`
	RelatedCommentID *int64      `json:"related_comment_id,omitempty"`
	IsRead           bool        `json:"is_read"`
	CreatedAt        string      `json:"created_at"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListRequest represents request for listing notifications
type ListRequest struct {
	Type     string `form:"type"`
	IsRead   *bool  `form:"is_read"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// ListResponse represents response for listing notifications
type ListResponse struct {
	List       []NotificationItem `json:"list"`
	Pagination Pagination         `json:"pagination"`
}

// List retrieves notifications for a user
func (s *Service) List(ctx context.Context, userID int64, req *ListRequest) (*ListResponse, error) {
	result, err := s.notifRepo.List(ctx, notification.ListParams{
		UserID:   userID,
		Type:     req.Type,
		IsRead:   req.IsRead,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// Get unique actor IDs
	actorIDs := make([]int64, 0, len(result.Notifications))
	actorIDMap := make(map[int64]bool)
	for _, n := range result.Notifications {
		if n.ActorID.Valid && !actorIDMap[n.ActorID.Int64] {
			actorIDs = append(actorIDs, n.ActorID.Int64)
			actorIDMap[n.ActorID.Int64] = true
		}
	}

	// Get actors
	actorsModelMap, err := s.notifRepo.GetUserMap(ctx, actorIDs)
	if err != nil {
		return nil, err
	}

	// Convert to UserBrief map
	actorsMap := make(map[int64]*UserBrief)
	for id, u := range actorsModelMap {
		ub := &UserBrief{
			ID:       u.ID,
			Username: u.Username,
		}
		if u.Avatar.Valid {
			ub.Avatar = &u.Avatar.String
		}
		actorsMap[id] = ub
	}

	// Build response
	list := make([]NotificationItem, len(result.Notifications))
	for i, n := range result.Notifications {
		item := NotificationItem{
			ID:        n.ID,
			Type:      n.Type,
			Title:     n.Title,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
		}

		if n.Content.Valid {
			item.Content = &n.Content.String
		}

		if n.ActorID.Valid {
			item.Actor = actorsMap[n.ActorID.Int64]
		}

		if n.RelatedPhotoID.Valid {
			item.RelatedPhotoID = &n.RelatedPhotoID.Int64
		}

		if n.RelatedCommentID.Valid {
			item.RelatedCommentID = &n.RelatedCommentID.Int64
		}

		list[i] = item
	}

	return &ListResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

// UnreadCountResponse represents response for unread count
type UnreadCountResponse struct {
	Count int `json:"count"`
}

// GetUnreadCount returns the unread notification count
func (s *Service) GetUnreadCount(ctx context.Context, userID int64) (*UnreadCountResponse, error) {
	count, err := s.notifRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &UnreadCountResponse{Count: count}, nil
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(ctx context.Context, userID, notifID int64) error {
	err := s.notifRepo.MarkAsRead(ctx, notifID, userID)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrNotificationNotFound
	}
	return err
}

// MarkAllAsRead marks all notifications as read
func (s *Service) MarkAllAsRead(ctx context.Context, userID int64) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

// ============================================
// Notification Creation Methods
// ============================================

// CreateLikeNotification creates a notification for a photo like
func (s *Service) CreateLikeNotification(ctx context.Context, photoOwnerID, actorID, photoID int64, photoTitle string) error {
	if photoOwnerID == actorID {
		return nil // Don't notify self
	}

	_, err := s.notifRepo.Create(ctx, &notification.Notification{
		UserID:         photoOwnerID,
		ActorID:        sql.NullInt64{Int64: actorID, Valid: true},
		Type:           string(notification.TypeLike),
		Title:          "有人点赞了你的照片",
		Content:        sql.NullString{String: photoTitle, Valid: true},
		RelatedPhotoID: sql.NullInt64{Int64: photoID, Valid: true},
	})
	return err
}

// CreateCommentNotification creates a notification for a photo comment
func (s *Service) CreateCommentNotification(ctx context.Context, photoOwnerID, actorID, photoID, commentID int64, photoTitle string) error {
	if photoOwnerID == actorID {
		return nil // Don't notify self
	}

	_, err := s.notifRepo.Create(ctx, &notification.Notification{
		UserID:           photoOwnerID,
		ActorID:          sql.NullInt64{Int64: actorID, Valid: true},
		Type:             string(notification.TypeComment),
		Title:            "有人评论了你的照片",
		Content:          sql.NullString{String: photoTitle, Valid: true},
		RelatedPhotoID:   sql.NullInt64{Int64: photoID, Valid: true},
		RelatedCommentID: sql.NullInt64{Int64: commentID, Valid: true},
	})
	return err
}

// CreateReplyNotification creates a notification for a comment reply
func (s *Service) CreateReplyNotification(ctx context.Context, commentOwnerID, actorID, photoID, commentID int64) error {
	if commentOwnerID == actorID {
		return nil // Don't notify self
	}

	_, err := s.notifRepo.Create(ctx, &notification.Notification{
		UserID:           commentOwnerID,
		ActorID:          sql.NullInt64{Int64: actorID, Valid: true},
		Type:             string(notification.TypeReply),
		Title:            "有人回复了你的评论",
		RelatedPhotoID:   sql.NullInt64{Int64: photoID, Valid: true},
		RelatedCommentID: sql.NullInt64{Int64: commentID, Valid: true},
	})
	return err
}

// CreateFeaturedNotification creates a notification for a featured photo
func (s *Service) CreateFeaturedNotification(ctx context.Context, photoOwnerID, photoID int64, photoTitle string) error {
	_, err := s.notifRepo.Create(ctx, &notification.Notification{
		UserID:         photoOwnerID,
		Type:           string(notification.TypeFeatured),
		Title:          "恭喜！你的照片入选精选",
		Content:        sql.NullString{String: photoTitle, Valid: true},
		RelatedPhotoID: sql.NullInt64{Int64: photoID, Valid: true},
	})
	return err
}

// CreateReviewNotification creates a notification for a photo review result
func (s *Service) CreateReviewNotification(ctx context.Context, photoOwnerID, photoID int64, approved bool, reason string) error {
	var title string
	if approved {
		title = "你的照片已通过审核"
	} else {
		title = "你的照片未通过审核"
	}

	var content sql.NullString
	if reason != "" {
		content = sql.NullString{String: reason, Valid: true}
	}

	_, err := s.notifRepo.Create(ctx, &notification.Notification{
		UserID:         photoOwnerID,
		Type:           string(notification.TypeReview),
		Title:          title,
		Content:        content,
		RelatedPhotoID: sql.NullInt64{Int64: photoID, Valid: true},
	})
	return err
}

// CreateSystemNotification creates a system notification
func (s *Service) CreateSystemNotification(ctx context.Context, userID int64, title, content string) error {
	var contentNull sql.NullString
	if content != "" {
		contentNull = sql.NullString{String: content, Valid: true}
	}

	_, err := s.notifRepo.Create(ctx, &notification.Notification{
		UserID:  userID,
		Type:    string(notification.TypeSystem),
		Title:   title,
		Content: contentNull,
	})
	return err
}
