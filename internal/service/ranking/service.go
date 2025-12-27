package ranking

import (
	"context"

	"QuanPhotos/internal/repository/postgresql/ranking"
)

// Service handles ranking business logic
type Service struct {
	rankingRepo *ranking.RankingRepository
	baseURL     string
}

// New creates a new ranking service
func New(rankingRepo *ranking.RankingRepository, baseURL string) *Service {
	return &Service{
		rankingRepo: rankingRepo,
		baseURL:     baseURL,
	}
}

// PhotoRankingRequest represents request for photo ranking
type PhotoRankingRequest struct {
	Period string `form:"period"` // day, week, month, all
	Limit  int    `form:"limit"`
	SortBy string `form:"sort_by"` // like_count, view_count, comment_count
}

// PhotoRankingItem represents a photo in ranking response
type PhotoRankingItem struct {
	Rank         int        `json:"rank"`
	ID           int64      `json:"id"`
	Title        string     `json:"title"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	LikeCount    int        `json:"like_count"`
	ViewCount    int        `json:"view_count"`
	CommentCount int        `json:"comment_count"`
	ShareCount   int        `json:"share_count"`
	User         *UserBrief `json:"user,omitempty"`
}

// UserBrief represents brief user info
type UserBrief struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

// PhotoRankingResponse represents response for photo ranking
type PhotoRankingResponse struct {
	Period string             `json:"period"`
	SortBy string             `json:"sort_by"`
	List   []PhotoRankingItem `json:"list"`
}

// GetPhotoRanking retrieves photo ranking
func (s *Service) GetPhotoRanking(ctx context.Context, req *PhotoRankingRequest) (*PhotoRankingResponse, error) {
	if req.Period == "" {
		req.Period = "all"
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "like_count"
	}

	items, err := s.rankingRepo.GetPhotoRanking(ctx, ranking.RankingParams{
		Period: req.Period,
		Limit:  req.Limit,
		SortBy: req.SortBy,
	})
	if err != nil {
		return nil, err
	}

	// Get unique user IDs
	userIDs := make([]int64, 0, len(items))
	userIDMap := make(map[int64]bool)
	for _, item := range items {
		if !userIDMap[item.UserID] {
			userIDs = append(userIDs, item.UserID)
			userIDMap[item.UserID] = true
		}
	}

	// Get users
	usersMap, err := s.rankingRepo.GetUserMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Build response
	list := make([]PhotoRankingItem, len(items))
	for i, item := range items {
		listItem := PhotoRankingItem{
			Rank:         i + 1,
			ID:           item.ID,
			Title:        item.Title,
			LikeCount:    item.LikeCount,
			ViewCount:    item.ViewCount,
			CommentCount: item.CommentCount,
			ShareCount:   item.ShareCount,
		}

		if item.ThumbnailPath != nil {
			url := s.baseURL + *item.ThumbnailPath
			listItem.ThumbnailURL = &url
		}

		if u, ok := usersMap[item.UserID]; ok {
			ub := &UserBrief{
				ID:       u.ID,
				Username: u.Username,
			}
			if u.Avatar.Valid {
				ub.Avatar = &u.Avatar.String
			}
			listItem.User = ub
		}

		list[i] = listItem
	}

	return &PhotoRankingResponse{
		Period: req.Period,
		SortBy: req.SortBy,
		List:   list,
	}, nil
}

// UserRankingRequest represents request for user ranking
type UserRankingRequest struct {
	Period string `form:"period"` // day, week, month, all
	Limit  int    `form:"limit"`
	SortBy string `form:"sort_by"` // photo_count, total_likes, total_views
}

// UserRankingItem represents a user in ranking response
type UserRankingItem struct {
	Rank       int     `json:"rank"`
	ID         int64   `json:"id"`
	Username   string  `json:"username"`
	Avatar     *string `json:"avatar,omitempty"`
	PhotoCount int     `json:"photo_count"`
	TotalLikes int     `json:"total_likes"`
	TotalViews int     `json:"total_views"`
}

// UserRankingResponse represents response for user ranking
type UserRankingResponse struct {
	Period string            `json:"period"`
	SortBy string            `json:"sort_by"`
	List   []UserRankingItem `json:"list"`
}

// GetUserRanking retrieves user ranking
func (s *Service) GetUserRanking(ctx context.Context, req *UserRankingRequest) (*UserRankingResponse, error) {
	if req.Period == "" {
		req.Period = "all"
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "total_likes"
	}

	items, err := s.rankingRepo.GetUserRanking(ctx, ranking.RankingParams{
		Period: req.Period,
		Limit:  req.Limit,
		SortBy: req.SortBy,
	})
	if err != nil {
		return nil, err
	}

	// Build response
	list := make([]UserRankingItem, len(items))
	for i, item := range items {
		listItem := UserRankingItem{
			Rank:       i + 1,
			ID:         item.ID,
			Username:   item.Username,
			PhotoCount: item.PhotoCount,
			TotalLikes: item.TotalLikes,
			TotalViews: item.TotalViews,
		}

		if item.Avatar != nil {
			listItem.Avatar = item.Avatar
		}

		list[i] = listItem
	}

	return &UserRankingResponse{
		Period: req.Period,
		SortBy: req.SortBy,
		List:   list,
	}, nil
}
