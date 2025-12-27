package handler

import (
	"errors"
	"net/http"
	"strconv"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/share"

	"github.com/gin-gonic/gin"
)

// ShareHandler handles share-related requests
type ShareHandler struct {
	shareService *share.Service
}

// NewShareHandler creates a new share handler
func NewShareHandler(shareService *share.Service) *ShareHandler {
	return &ShareHandler{
		shareService: shareService,
	}
}

// Share godoc
// @Summary Share a photo
// @Description Share a photo to a platform
// @Tags Share
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Param request body share.ShareRequest true "Share data"
// @Success 200 {object} response.Response{data=share.ShareResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id}/share [post]
func (h *ShareHandler) Share(c *gin.Context) {
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid photo id")
		return
	}

	userID := c.GetInt64("userID")

	var req share.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.PhotoID = photoID

	result, err := h.shareService.Share(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, share.ErrPhotoNotFound) {
			response.NotFound(c, "photo not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}
