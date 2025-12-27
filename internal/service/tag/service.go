package tag

import (
	"context"
	"errors"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/tag"
)

var (
	ErrTagNotFound   = errors.New("tag not found")
	ErrDuplicateName = errors.New("tag name already exists")
)

// Service handles tag business logic
type Service struct {
	tagRepo *tag.TagRepository
	baseURL string
}

// New creates a new tag service
func New(tagRepo *tag.TagRepository, baseURL string) *Service {
	return &Service{
		tagRepo: tagRepo,
		baseURL: baseURL,
	}
}

// ListRequest represents request for listing tags
type ListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	OrderBy  string `form:"order_by"` // photo_count, name, created_at
}

// TagItem represents a tag in response
type TagItem struct {
	ID         int32  `json:"id"`
	Name       string `json:"name"`
	PhotoCount int    `json:"photo_count"`
	CreatedAt  string `json:"created_at"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResponse represents response for listing tags
type ListResponse struct {
	List       []TagItem  `json:"list"`
	Pagination Pagination `json:"pagination"`
}

// List retrieves popular tags
func (s *Service) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	result, err := s.tagRepo.List(ctx, tag.ListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		OrderBy:  req.OrderBy,
	})
	if err != nil {
		return nil, err
	}

	list := make([]TagItem, len(result.Tags))
	for i, t := range result.Tags {
		list[i] = s.toTagItem(t)
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

// SearchRequest represents request for searching tags
type SearchRequest struct {
	Keyword string `form:"q" binding:"required,min=1"`
	Limit   int    `form:"limit"`
}

// Search searches tags by keyword
func (s *Service) Search(ctx context.Context, req *SearchRequest) ([]TagItem, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}

	tags, err := s.tagRepo.Search(ctx, req.Keyword, req.Limit)
	if err != nil {
		return nil, err
	}

	result := make([]TagItem, len(tags))
	for i, t := range tags {
		result[i] = s.toTagItem(t)
	}

	return result, nil
}

// GetByID retrieves a tag by ID
func (s *Service) GetByID(ctx context.Context, id int32) (*TagItem, error) {
	t, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	item := s.toTagItem(t)
	return &item, nil
}

// ListPhotosRequest represents request for listing photos with a tag
type ListPhotosRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// PhotoListItem represents a photo in list
type PhotoListItem struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	ThumbnailURL  string     `json:"thumbnail_url"`
	AircraftType  *string    `json:"aircraft_type,omitempty"`
	Airline       *string    `json:"airline,omitempty"`
	Registration  *string    `json:"registration,omitempty"`
	ViewCount     int        `json:"view_count"`
	LikeCount     int        `json:"like_count"`
	FavoriteCount int        `json:"favorite_count"`
	CreatedAt     string     `json:"created_at"`
	User          *UserBrief `json:"user"`
}

// UserBrief represents brief user info
type UserBrief struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

// ListPhotosResponse represents response for listing photos
type ListPhotosResponse struct {
	Tag        TagItem         `json:"tag"`
	List       []PhotoListItem `json:"list"`
	Pagination Pagination      `json:"pagination"`
}

// ListPhotos retrieves photos with a specific tag
func (s *Service) ListPhotos(ctx context.Context, tagID int32, req *ListPhotosRequest) (*ListPhotosResponse, error) {
	// Get tag info
	t, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	// Get photos
	result, err := s.tagRepo.ListPhotos(ctx, tag.ListPhotosParams{
		TagID:     tagID,
		Page:      req.Page,
		PageSize:  req.PageSize,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	// Get user IDs
	userIDs := make([]int64, 0, len(result.Photos))
	userIDMap := make(map[int64]bool)
	for _, p := range result.Photos {
		if !userIDMap[p.UserID] {
			userIDs = append(userIDs, p.UserID)
			userIDMap[p.UserID] = true
		}
	}

	// Get users
	usersMap, err := s.tagRepo.GetUserMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Build response
	list := make([]PhotoListItem, len(result.Photos))
	for i, p := range result.Photos {
		item := PhotoListItem{
			ID:            p.ID,
			Title:         p.Title,
			ViewCount:     p.ViewCount,
			LikeCount:     p.LikeCount,
			FavoriteCount: p.FavoriteCount,
			CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		}

		if p.ThumbnailPath.Valid {
			item.ThumbnailURL = s.baseURL + p.ThumbnailPath.String
		}
		if p.AircraftType.Valid {
			item.AircraftType = &p.AircraftType.String
		}
		if p.Airline.Valid {
			item.Airline = &p.Airline.String
		}
		if p.Registration.Valid {
			item.Registration = &p.Registration.String
		}

		if u, ok := usersMap[p.UserID]; ok {
			item.User = &UserBrief{
				ID:       u.ID,
				Username: u.Username,
			}
			if u.Avatar.Valid {
				item.User.Avatar = &u.Avatar.String
			}
		}

		list[i] = item
	}

	return &ListPhotosResponse{
		Tag:  s.toTagItem(t),
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

func (s *Service) toTagItem(t *model.Tag) TagItem {
	return TagItem{
		ID:         t.ID,
		Name:       t.Name,
		PhotoCount: t.PhotoCount,
		CreatedAt:  t.CreatedAt.Format(time.RFC3339),
	}
}
