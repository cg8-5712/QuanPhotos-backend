package category

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/category"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrDuplicateName    = errors.New("category name already exists")
	ErrHasPhotos        = errors.New("category has photos, cannot delete")
)

// Service handles category business logic
type Service struct {
	categoryRepo *category.CategoryRepository
	baseURL      string
}

// New creates a new category service
func New(categoryRepo *category.CategoryRepository, baseURL string) *Service {
	return &Service{
		categoryRepo: categoryRepo,
		baseURL:      baseURL,
	}
}

// ListRequest represents request for listing categories
type ListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// CategoryItem represents a category in response
type CategoryItem struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	NameEN      string  `json:"name_en"`
	Description *string `json:"description,omitempty"`
	SortOrder   int     `json:"sort_order"`
	PhotoCount  int     `json:"photo_count"`
	CreatedAt   string  `json:"created_at"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResponse represents response for listing categories
type ListResponse struct {
	List       []CategoryItem `json:"list"`
	Pagination Pagination     `json:"pagination"`
}

// List retrieves all categories
func (s *Service) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	result, err := s.categoryRepo.List(ctx, category.ListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]CategoryItem, len(result.Categories))
	for i, c := range result.Categories {
		list[i] = s.toCategoryItem(c)
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

// GetByID retrieves a category by ID
func (s *Service) GetByID(ctx context.Context, id int32) (*CategoryItem, error) {
	cat, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	item := s.toCategoryItem(cat)
	return &item, nil
}

// CreateRequest represents request for creating a category
type CreateRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	NameEN      string `json:"name_en" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	SortOrder   int    `json:"sort_order"`
}

// Create creates a new category
func (s *Service) Create(ctx context.Context, req *CreateRequest) (*CategoryItem, error) {
	cat, err := s.categoryRepo.Create(ctx, req.Name, req.NameEN, req.Description, req.SortOrder)
	if err != nil {
		if errors.Is(err, postgresql.ErrDuplicateKey) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}

	item := s.toCategoryItem(cat)
	return &item, nil
}

// UpdateRequest represents request for updating a category
type UpdateRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	NameEN      string `json:"name_en" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	SortOrder   int    `json:"sort_order"`
}

// Update updates a category
func (s *Service) Update(ctx context.Context, id int32, req *UpdateRequest) (*CategoryItem, error) {
	cat, err := s.categoryRepo.Update(ctx, id, req.Name, req.NameEN, req.Description, req.SortOrder)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		if errors.Is(err, postgresql.ErrDuplicateKey) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}

	item := s.toCategoryItem(cat)
	return &item, nil
}

// Delete deletes a category
func (s *Service) Delete(ctx context.Context, id int32, force bool) error {
	// Check if category has photos
	if !force {
		count, err := s.categoryRepo.GetPhotoCount(ctx, id)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrHasPhotos
		}
	}

	err := s.categoryRepo.Delete(ctx, id)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrCategoryNotFound
	}
	return err
}

// ListPhotosRequest represents request for listing photos in a category
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
	List       []PhotoListItem `json:"list"`
	Pagination Pagination      `json:"pagination"`
}

func (s *Service) toCategoryItem(c *model.Category) CategoryItem {
	item := CategoryItem{
		ID:         c.ID,
		Name:       c.Name,
		NameEN:     c.NameEN,
		SortOrder:  c.SortOrder,
		PhotoCount: c.PhotoCount,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.Description.Valid {
		item.Description = &c.Description.String
	}
	return item
}
