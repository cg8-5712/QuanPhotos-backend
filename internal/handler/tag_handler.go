package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/tag"
)

// TagHandler handles tag HTTP requests
type TagHandler struct {
	tagService *tag.Service
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagService *tag.Service) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// List lists popular tags
// @Summary List tags
// @Description Get a list of popular tags
// @Tags Tags
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Param order_by query string false "Order by: photo_count, name, created_at" default(photo_count)
// @Success 200 {object} response.Response
// @Router /api/v1/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	var req tag.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}
	if req.OrderBy == "" {
		req.OrderBy = "photo_count"
	}

	result, err := h.tagService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to list tags")
		return
	}

	response.Success(c, result)
}

// Search searches tags by keyword
// @Summary Search tags
// @Description Search tags by keyword
// @Tags Tags
// @Produce json
// @Param q query string true "Search keyword"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {object} response.Response
// @Router /api/v1/tags/search [get]
func (h *TagHandler) Search(c *gin.Context) {
	var req tag.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Search keyword is required")
		return
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	result, err := h.tagService.Search(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to search tags")
		return
	}

	response.Success(c, result)
}

// ListPhotos lists photos with a specific tag
// @Summary List photos by tag
// @Description Get photos with a specific tag
// @Tags Tags
// @Produce json
// @Param id path int true "Tag ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort by: created_at, like_count, view_count" default(created_at)
// @Param sort_order query string false "Sort order: asc, desc" default(desc)
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/tags/{id}/photos [get]
func (h *TagHandler) ListPhotos(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid tag ID")
		return
	}

	var req tag.ListPhotosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	result, err := h.tagService.ListPhotos(c.Request.Context(), int32(id), &req)
	if err != nil {
		if errors.Is(err, tag.ErrTagNotFound) {
			response.NotFound(c, "Tag not found")
			return
		}
		response.InternalError(c, "Failed to list photos")
		return
	}

	response.Success(c, result)
}
