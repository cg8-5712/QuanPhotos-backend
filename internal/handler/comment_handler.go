package handler

import (
	"errors"
	"net/http"
	"strconv"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/comment"

	"github.com/gin-gonic/gin"
)

// CommentHandler handles comment-related requests
type CommentHandler struct {
	commentService *comment.Service
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentService *comment.Service) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// List godoc
// @Summary Get comments for a photo
// @Description Get paginated comments for a specific photo
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Photo ID"
// @Param parent_id query int false "Parent comment ID (for replies)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort by field" Enums(created_at, like_count)
// @Success 200 {object} response.Response{data=comment.ListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/photos/{id}/comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid photo id")
		return
	}

	var req comment.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.PhotoID = photoID

	// Get current user ID if authenticated
	var currentUserID *int64
	if userID, exists := c.Get("userID"); exists {
		id := userID.(int64)
		currentUserID = &id
	}

	result, err := h.commentService.List(c.Request.Context(), &req, currentUserID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Create godoc
// @Summary Create a comment
// @Description Create a new comment on a photo
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Param request body comment.CreateRequest true "Comment data"
// @Success 201 {object} response.Response{data=comment.CommentItem}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid photo id")
		return
	}

	userID := c.GetInt64("userID")

	var req comment.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.PhotoID = photoID

	result, err := h.commentService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, comment.ErrPhotoNotFound) {
			response.NotFound(c, "photo not found")
			return
		}
		if errors.Is(err, comment.ErrParentNotFound) {
			response.NotFound(c, "parent comment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

// Delete godoc
// @Summary Delete a comment
// @Description Delete a comment (owner or admin only)
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	userID := c.GetInt64("userID")
	role := c.GetString("role")
	isAdmin := role == "admin" || role == "superadmin"

	err = h.commentService.Delete(c.Request.Context(), commentID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			response.NotFound(c, "comment not found")
			return
		}
		if errors.Is(err, comment.ErrNotOwner) {
			response.Forbidden(c, "you are not the owner of this comment")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// AddLike godoc
// @Summary Like a comment
// @Description Add a like to a comment
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/comments/{id}/like [post]
func (h *CommentHandler) AddLike(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	userID := c.GetInt64("userID")

	err = h.commentService.AddLike(c.Request.Context(), userID, commentID)
	if err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			response.NotFound(c, "comment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// RemoveLike godoc
// @Summary Unlike a comment
// @Description Remove a like from a comment
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/comments/{id}/like [delete]
func (h *CommentHandler) RemoveLike(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	userID := c.GetInt64("userID")

	err = h.commentService.RemoveLike(c.Request.Context(), userID, commentID)
	if err != nil {
		if errors.Is(err, comment.ErrNotLiked) {
			response.BadRequest(c, "comment is not liked")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
