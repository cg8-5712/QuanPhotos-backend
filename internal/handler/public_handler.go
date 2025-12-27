package handler

import (
	"errors"
	"strconv"
	"time"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/photo"
	"QuanPhotos/internal/service/ranking"

	"github.com/gin-gonic/gin"
)

// PublicHandler handles public-facing requests
type PublicHandler struct {
	photoRepo      *photo.PhotoRepository
	rankingService *ranking.Service
	baseURL        string
}

// NewPublicHandler creates a new public handler
func NewPublicHandler(photoRepo *photo.PhotoRepository, rankingService *ranking.Service, baseURL string) *PublicHandler {
	return &PublicHandler{
		photoRepo:      photoRepo,
		rankingService: rankingService,
		baseURL:        baseURL,
	}
}

// FeaturedPhotoItem represents a featured photo in response
type FeaturedPhotoItem struct {
	ID           int64   `json:"id"`
	Title        string  `json:"title"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	LikeCount    int     `json:"like_count"`
	ViewCount    int     `json:"view_count"`
	UserID       int64   `json:"user_id"`
}

// FeaturedListRequest represents request for listing featured photos
type FeaturedListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// FeaturedListResponse represents response for listing featured photos
type FeaturedListResponse struct {
	List       []FeaturedPhotoItem `json:"list"`
	Pagination Pagination          `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListFeatured godoc
// @Summary Get featured photos
// @Description Get paginated list of featured photos
// @Tags Featured
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=FeaturedListResponse}
// @Router /api/v1/featured [get]
func (h *PublicHandler) ListFeatured(c *gin.Context) {
	var req FeaturedListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.photoRepo.ListFeaturedPhotos(c.Request.Context(), photo.FeaturedListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	list := make([]FeaturedPhotoItem, len(result.Photos))
	for i, p := range result.Photos {
		item := FeaturedPhotoItem{
			ID:        p.ID,
			Title:     p.Title,
			LikeCount: p.LikeCount,
			ViewCount: p.ViewCount,
			UserID:    p.UserID,
		}
		if p.ThumbnailPath.Valid {
			url := h.baseURL + p.ThumbnailPath.String
			item.ThumbnailURL = &url
		}
		list[i] = item
	}

	response.Success(c, FeaturedListResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	})
}

// PhotoRanking godoc
// @Summary Get photo ranking
// @Description Get photo ranking based on likes, views, or comments
// @Tags Rankings
// @Accept json
// @Produce json
// @Param period query string false "Time period" Enums(day, week, month, all) default(all)
// @Param limit query int false "Number of items" default(20)
// @Param sort_by query string false "Sort by field" Enums(like_count, view_count, comment_count) default(like_count)
// @Success 200 {object} response.Response{data=ranking.PhotoRankingResponse}
// @Router /api/v1/rankings/photos [get]
func (h *PublicHandler) PhotoRanking(c *gin.Context) {
	var req ranking.PhotoRankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.rankingService.GetPhotoRanking(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// UserRanking godoc
// @Summary Get user ranking
// @Description Get user ranking based on photo count or total likes
// @Tags Rankings
// @Accept json
// @Produce json
// @Param period query string false "Time period" Enums(day, week, month, all) default(all)
// @Param limit query int false "Number of items" default(20)
// @Param sort_by query string false "Sort by field" Enums(photo_count, total_likes, total_views) default(total_likes)
// @Success 200 {object} response.Response{data=ranking.UserRankingResponse}
// @Router /api/v1/rankings/users [get]
func (h *PublicHandler) UserRanking(c *gin.Context) {
	var req ranking.UserRankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.rankingService.GetUserRanking(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// AnnouncementListRequest represents request for listing announcements
type AnnouncementListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// AnnouncementItem represents an announcement in response
type AnnouncementItem struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Summary     *string `json:"summary,omitempty"`
	IsPinned    bool    `json:"is_pinned"`
	PublishedAt *string `json:"published_at,omitempty"`
}

// AnnouncementListResponse represents response for listing announcements
type AnnouncementListResponse struct {
	List       []AnnouncementItem `json:"list"`
	Pagination Pagination         `json:"pagination"`
}

// ListAnnouncements godoc
// @Summary Get public announcements
// @Description Get paginated list of published announcements
// @Tags Announcements
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=AnnouncementListResponse}
// @Router /api/v1/announcements [get]
func (h *PublicHandler) ListAnnouncements(c *gin.Context) {
	var req AnnouncementListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.photoRepo.ListPublicAnnouncements(c.Request.Context(), photo.AnnouncementListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	list := make([]AnnouncementItem, len(result.Announcements))
	for i, a := range result.Announcements {
		item := AnnouncementItem{
			ID:       a.ID,
			Title:    a.Title,
			IsPinned: a.IsPinned,
		}
		if a.Summary.Valid {
			item.Summary = &a.Summary.String
		}
		if a.PublishedAt.Valid {
			publishedAt := a.PublishedAt.Time.Format(time.RFC3339)
			item.PublishedAt = &publishedAt
		}
		list[i] = item
	}

	response.Success(c, AnnouncementListResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	})
}

// AnnouncementDetailResponse represents detailed announcement response
type AnnouncementDetailResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Summary     *string `json:"summary,omitempty"`
	Content     string  `json:"content"`
	IsPinned    bool    `json:"is_pinned"`
	PublishedAt *string `json:"published_at,omitempty"`
}

// GetAnnouncement godoc
// @Summary Get announcement detail
// @Description Get a published announcement by ID
// @Tags Announcements
// @Accept json
// @Produce json
// @Param id path int true "Announcement ID"
// @Success 200 {object} response.Response{data=AnnouncementDetailResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/announcements/{id} [get]
func (h *PublicHandler) GetAnnouncement(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid announcement id")
		return
	}

	ann, err := h.photoRepo.GetPublicAnnouncementByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			response.NotFound(c, "announcement not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	resp := AnnouncementDetailResponse{
		ID:       ann.ID,
		Title:    ann.Title,
		Content:  ann.Content,
		IsPinned: ann.IsPinned,
	}
	if ann.Summary.Valid {
		resp.Summary = &ann.Summary.String
	}
	if ann.PublishedAt.Valid {
		publishedAt := ann.PublishedAt.Time.Format(time.RFC3339)
		resp.PublishedAt = &publishedAt
	}

	response.Success(c, resp)
}
