package comment

import (
	"context"
	"errors"
	"time"

	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/comment"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrPhotoNotFound   = errors.New("photo not found")
	ErrNotOwner        = errors.New("you are not the owner of this comment")
	ErrNotLiked        = errors.New("comment is not liked")
	ErrParentNotFound  = errors.New("parent comment not found")
)

// Service handles comment business logic
type Service struct {
	commentRepo *comment.CommentRepository
	baseURL     string
}

// New creates a new comment service
func New(commentRepo *comment.CommentRepository, baseURL string) *Service {
	return &Service{
		commentRepo: commentRepo,
		baseURL:     baseURL,
	}
}

// ListRequest represents request for listing comments
type ListRequest struct {
	PhotoID  int64  `form:"-"`
	ParentID *int64 `form:"parent_id"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	SortBy   string `form:"sort_by"` // created_at, like_count
}

// UserBrief represents brief user info
type UserBrief struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

// CommentItem represents a comment in response
type CommentItem struct {
	ID         int64         `json:"id"`
	PhotoID    int64         `json:"photo_id"`
	ParentID   *int64        `json:"parent_id,omitempty"`
	Content    string        `json:"content"`
	LikeCount  int           `json:"like_count"`
	ReplyCount int           `json:"reply_count"`
	IsLiked    bool          `json:"is_liked"`
	CreatedAt  string        `json:"created_at"`
	User       *UserBrief    `json:"user"`
	Replies    []CommentItem `json:"replies,omitempty"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResponse represents response for listing comments
type ListResponse struct {
	List       []CommentItem `json:"list"`
	Pagination Pagination    `json:"pagination"`
}

// List retrieves comments for a photo
func (s *Service) List(ctx context.Context, req *ListRequest, currentUserID *int64) (*ListResponse, error) {
	result, err := s.commentRepo.List(ctx, comment.ListParams{
		PhotoID:  req.PhotoID,
		ParentID: req.ParentID,
		Page:     req.Page,
		PageSize: req.PageSize,
		SortBy:   req.SortBy,
	})
	if err != nil {
		return nil, err
	}

	// Get unique user IDs
	userIDs := make([]int64, 0, len(result.Comments))
	userIDMap := make(map[int64]bool)
	for _, c := range result.Comments {
		if !userIDMap[c.UserID] {
			userIDs = append(userIDs, c.UserID)
			userIDMap[c.UserID] = true
		}
	}

	// Get users
	usersModelMap, err := s.commentRepo.GetUserMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Convert to UserBrief map
	usersMap := make(map[int64]*UserBrief)
	for id, u := range usersModelMap {
		ub := &UserBrief{
			ID:       u.ID,
			Username: u.Username,
		}
		if u.Avatar.Valid {
			ub.Avatar = &u.Avatar.String
		}
		usersMap[id] = ub
	}

	// Build response
	list := make([]CommentItem, len(result.Comments))
	for i, c := range result.Comments {
		item := s.toCommentItem(c, usersMap)

		// Check if liked by current user
		if currentUserID != nil {
			item.IsLiked, _ = s.commentRepo.IsLiked(ctx, *currentUserID, c.ID)
		}

		// Get replies if this is a top-level comment
		if req.ParentID == nil && c.ReplyCount > 0 {
			replies, err := s.commentRepo.GetReplies(ctx, c.ID, 3)
			if err == nil && len(replies) > 0 {
				// Get reply user IDs
				for _, r := range replies {
					if !userIDMap[r.UserID] {
						userIDs = append(userIDs, r.UserID)
						userIDMap[r.UserID] = true
					}
				}
				// Get additional users if needed
				if len(userIDs) > len(usersMap) {
					additionalUsers, _ := s.commentRepo.GetUserMap(ctx, userIDs)
					for id, u := range additionalUsers {
						if _, exists := usersMap[id]; !exists {
							ub := &UserBrief{
								ID:       u.ID,
								Username: u.Username,
							}
							if u.Avatar.Valid {
								ub.Avatar = &u.Avatar.String
							}
							usersMap[id] = ub
						}
					}
				}

				item.Replies = make([]CommentItem, len(replies))
				for j, r := range replies {
					replyItem := s.toCommentItem(r, usersMap)
					if currentUserID != nil {
						replyItem.IsLiked, _ = s.commentRepo.IsLiked(ctx, *currentUserID, r.ID)
					}
					item.Replies[j] = replyItem
				}
			}
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

// CreateRequest represents request for creating a comment
type CreateRequest struct {
	PhotoID  int64  `json:"-"`
	ParentID *int64 `json:"parent_id"`
	Content  string `json:"content" binding:"required,min=1,max=1000"`
}

// Create creates a new comment
func (s *Service) Create(ctx context.Context, userID int64, req *CreateRequest) (*CommentItem, error) {
	// Check if photo exists
	exists, err := s.commentRepo.PhotoExists(ctx, req.PhotoID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPhotoNotFound
	}

	// Check if parent comment exists (if provided)
	if req.ParentID != nil {
		parentExists, err := s.commentRepo.Exists(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if !parentExists {
			return nil, ErrParentNotFound
		}
	}

	// Create comment
	c, err := s.commentRepo.Create(ctx, req.PhotoID, userID, req.ParentID, req.Content)
	if err != nil {
		return nil, err
	}

	// Get user info
	usersModelMap, err := s.commentRepo.GetUserMap(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}

	// Convert to UserBrief map
	usersMap := make(map[int64]*UserBrief)
	for id, u := range usersModelMap {
		ub := &UserBrief{
			ID:       u.ID,
			Username: u.Username,
		}
		if u.Avatar.Valid {
			ub.Avatar = &u.Avatar.String
		}
		usersMap[id] = ub
	}

	item := s.toCommentItem(c, usersMap)
	return &item, nil
}

// Delete deletes a comment
func (s *Service) Delete(ctx context.Context, commentID, userID int64, isAdmin bool) error {
	// Check if comment exists
	exists, err := s.commentRepo.Exists(ctx, commentID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCommentNotFound
	}

	// Check ownership (unless admin)
	if !isAdmin {
		isOwner, err := s.commentRepo.IsOwnedBy(ctx, commentID, userID)
		if err != nil {
			return err
		}
		if !isOwner {
			return ErrNotOwner
		}
	}

	return s.commentRepo.Delete(ctx, commentID)
}

// AddLike adds a like to a comment
func (s *Service) AddLike(ctx context.Context, userID, commentID int64) error {
	// Check if comment exists
	exists, err := s.commentRepo.Exists(ctx, commentID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCommentNotFound
	}

	return s.commentRepo.AddLike(ctx, userID, commentID)
}

// RemoveLike removes a like from a comment
func (s *Service) RemoveLike(ctx context.Context, userID, commentID int64) error {
	err := s.commentRepo.RemoveLike(ctx, userID, commentID)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrNotLiked
	}
	return err
}

func (s *Service) toCommentItem(c *comment.Comment, usersMap map[int64]*UserBrief) CommentItem {
	item := CommentItem{
		ID:         c.ID,
		PhotoID:    c.PhotoID,
		Content:    c.Content,
		LikeCount:  c.LikeCount,
		ReplyCount: c.ReplyCount,
		CreatedAt:  c.CreatedAt.Format(time.RFC3339),
	}

	if c.ParentID.Valid {
		item.ParentID = &c.ParentID.Int64
	}

	if u, ok := usersMap[c.UserID]; ok {
		item.User = u
	}

	return item
}
