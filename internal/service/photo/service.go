package photo

import (
	"context"
	"errors"

	"QuanPhotos/internal/config"
	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/storage"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/photo"
)

var (
	ErrPhotoNotFound = errors.New("photo not found")
	ErrNotOwner      = errors.New("you are not the owner of this photo")
	ErrNotFavorited  = errors.New("photo is not in favorites")
	ErrNotLiked      = errors.New("photo is not liked")
)

// Service handles photo business logic
type Service struct {
	photoRepo *photo.PhotoRepository
	uploader  *Uploader
	baseURL   string
}

// New creates a new photo service
func New(photoRepo *photo.PhotoRepository, baseURL string) *Service {
	return &Service{
		photoRepo: photoRepo,
		baseURL:   baseURL,
	}
}

// NewWithUploader creates a new photo service with uploader support
func NewWithUploader(photoRepo *photo.PhotoRepository, localStorage *storage.LocalStorage, cfg *config.Config) *Service {
	return &Service{
		photoRepo: photoRepo,
		uploader:  NewUploader(localStorage, photoRepo, cfg),
		baseURL:   cfg.Storage.BaseURL,
	}
}

// Upload uploads a new photo
func (s *Service) Upload(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	if s.uploader == nil {
		return nil, errors.New("uploader not initialized")
	}
	return s.uploader.Upload(ctx, req)
}

// ListRequest represents request for listing photos
type ListRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	CategoryID   int32  `form:"category_id"`
	AircraftType string `form:"aircraft_type"`
	Airline      string `form:"airline"`
	Airport      string `form:"airport"`
	Keyword      string `form:"keyword"`
	SortBy       string `form:"sort_by"`
	SortOrder    string `form:"sort_order"`
}

// ListResponse represents response for listing photos
type ListResponse struct {
	List       []*model.PhotoListItem `json:"list"`
	Pagination Pagination             `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// List retrieves a paginated list of photos
func (s *Service) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	result, err := s.photoRepo.List(ctx, photo.ListParams{
		Page:         req.Page,
		PageSize:     req.PageSize,
		CategoryID:   req.CategoryID,
		AircraftType: req.AircraftType,
		Airline:      req.Airline,
		Airport:      req.Airport,
		Keyword:      req.Keyword,
		SortBy:       req.SortBy,
		SortOrder:    req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	// Get unique user IDs
	userIDs := make([]int64, 0, len(result.Photos))
	userIDMap := make(map[int64]bool)
	for _, p := range result.Photos {
		if !userIDMap[p.UserID] {
			userIDs = append(userIDs, p.UserID)
			userIDMap[p.UserID] = true
		}
	}

	// Get users
	users, err := s.photoRepo.GetUserMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Build response
	list := make([]*model.PhotoListItem, len(result.Photos))
	for i, p := range result.Photos {
		var userBrief *model.UserBrief
		if u, ok := users[p.UserID]; ok {
			userBrief = &model.UserBrief{
				ID:       u.ID,
				Username: u.Username,
			}
			if u.Avatar.Valid {
				userBrief.Avatar = &u.Avatar.String
			}
		}
		list[i] = p.ToListItem(userBrief, s.baseURL)
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

// GetDetail retrieves photo detail
func (s *Service) GetDetail(ctx context.Context, photoID int64, currentUserID *int64) (*model.PhotoDetail, error) {
	// Get photo
	p, err := s.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrPhotoNotFound
		}
		return nil, err
	}

	// Check access permission based on photo status
	// Only approved photos are publicly visible
	// Owner can see their own photos regardless of status
	if p.Status != model.PhotoStatusApproved {
		if currentUserID == nil || *currentUserID != p.UserID {
			return nil, ErrPhotoNotFound // Hide non-approved photos from non-owners
		}
	}

	// Increment view count
	_ = s.photoRepo.IncrementViewCount(ctx, photoID)

	// Get user
	user, err := s.photoRepo.GetUserByPhotoID(ctx, photoID)
	if err != nil {
		return nil, err
	}

	userBrief := &model.UserBrief{
		ID:       user.ID,
		Username: user.Username,
	}
	if user.Avatar.Valid {
		userBrief.Avatar = &user.Avatar.String
	}

	// Get category
	var categoryBrief *model.CategoryBrief
	if p.CategoryID.Valid {
		category, err := s.photoRepo.GetCategoryByID(ctx, p.CategoryID.Int32)
		if err == nil {
			categoryBrief = &model.CategoryBrief{
				ID:   category.ID,
				Name: category.Name,
			}
		}
	}

	// Get tags
	tags, err := s.photoRepo.GetTagsByPhotoID(ctx, photoID)
	if err != nil {
		tags = []string{}
	}

	// Check if favorited/liked by current user
	var isFavorited, isLiked bool
	if currentUserID != nil {
		isFavorited, _ = s.photoRepo.IsFavorited(ctx, *currentUserID, photoID)
		isLiked, _ = s.photoRepo.IsLiked(ctx, *currentUserID, photoID)
	}

	return p.ToDetail(userBrief, categoryBrief, tags, s.baseURL, isFavorited, isLiked), nil
}

// ListMyPhotos lists current user's photos
func (s *Service) ListMyPhotos(ctx context.Context, userID int64, page, pageSize int, status string) (*ListResponse, error) {
	params := photo.ListParams{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
		Status:   status,
	}
	if status == "" {
		params.Status = "all" // Show all statuses for own photos
	}

	result, err := s.photoRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// Get user
	users, err := s.photoRepo.GetUserMap(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}

	var userBrief *model.UserBrief
	if u, ok := users[userID]; ok {
		userBrief = &model.UserBrief{
			ID:       u.ID,
			Username: u.Username,
		}
		if u.Avatar.Valid {
			userBrief.Avatar = &u.Avatar.String
		}
	}

	// Build response
	list := make([]*model.PhotoListItem, len(result.Photos))
	for i, p := range result.Photos {
		list[i] = p.ToListItem(userBrief, s.baseURL)
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

// ListUserPhotos lists a user's approved photos
func (s *Service) ListUserPhotos(ctx context.Context, userID int64, page, pageSize int) (*ListResponse, error) {
	result, err := s.photoRepo.List(ctx, photo.ListParams{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
		Status:   string(model.PhotoStatusApproved),
	})
	if err != nil {
		return nil, err
	}

	// Get user
	users, err := s.photoRepo.GetUserMap(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}

	var userBrief *model.UserBrief
	if u, ok := users[userID]; ok {
		userBrief = &model.UserBrief{
			ID:       u.ID,
			Username: u.Username,
		}
		if u.Avatar.Valid {
			userBrief.Avatar = &u.Avatar.String
		}
	}

	// Build response
	list := make([]*model.PhotoListItem, len(result.Photos))
	for i, p := range result.Photos {
		list[i] = p.ToListItem(userBrief, s.baseURL)
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

// ListFavorites lists user's favorite photos
func (s *Service) ListFavorites(ctx context.Context, userID int64, page, pageSize int) (*ListResponse, error) {
	result, err := s.photoRepo.ListUserFavorites(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Get unique user IDs
	userIDs := make([]int64, 0, len(result.Photos))
	userIDMap := make(map[int64]bool)
	for _, p := range result.Photos {
		if !userIDMap[p.UserID] {
			userIDs = append(userIDs, p.UserID)
			userIDMap[p.UserID] = true
		}
	}

	// Get users
	users, err := s.photoRepo.GetUserMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Build response
	list := make([]*model.PhotoListItem, len(result.Photos))
	for i, p := range result.Photos {
		var userBrief *model.UserBrief
		if u, ok := users[p.UserID]; ok {
			userBrief = &model.UserBrief{
				ID:       u.ID,
				Username: u.Username,
			}
			if u.Avatar.Valid {
				userBrief.Avatar = &u.Avatar.String
			}
		}
		list[i] = p.ToListItem(userBrief, s.baseURL)
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

// AddFavorite adds a photo to user's favorites
func (s *Service) AddFavorite(ctx context.Context, userID, photoID int64) error {
	// Check if photo exists
	exists, err := s.photoRepo.Exists(ctx, photoID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPhotoNotFound
	}

	return s.photoRepo.AddFavorite(ctx, userID, photoID)
}

// RemoveFavorite removes a photo from user's favorites
func (s *Service) RemoveFavorite(ctx context.Context, userID, photoID int64) error {
	err := s.photoRepo.RemoveFavorite(ctx, userID, photoID)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrNotFavorited
	}
	return err
}

// AddLike adds a like to a photo
func (s *Service) AddLike(ctx context.Context, userID, photoID int64) error {
	// Check if photo exists
	exists, err := s.photoRepo.Exists(ctx, photoID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPhotoNotFound
	}

	return s.photoRepo.AddLike(ctx, userID, photoID)
}

// RemoveLike removes a like from a photo
func (s *Service) RemoveLike(ctx context.Context, userID, photoID int64) error {
	err := s.photoRepo.RemoveLike(ctx, userID, photoID)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrNotLiked
	}
	return err
}

// Delete deletes a photo
func (s *Service) Delete(ctx context.Context, photoID, userID int64, isAdmin bool) error {
	// Check if photo exists
	exists, err := s.photoRepo.Exists(ctx, photoID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPhotoNotFound
	}

	// Check ownership (unless admin)
	if !isAdmin {
		isOwner, err := s.photoRepo.IsOwnedBy(ctx, photoID, userID)
		if err != nil {
			return err
		}
		if !isOwner {
			return ErrNotOwner
		}
	}

	// TODO: Delete actual files from storage

	return s.photoRepo.Delete(ctx, photoID)
}
