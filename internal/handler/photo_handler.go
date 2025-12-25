package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/middleware"
	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/pkg/storage"
	"QuanPhotos/internal/service/photo"
)

// PhotoHandler handles photo HTTP requests
type PhotoHandler struct {
	photoService  *photo.Service
	maxUploadSize int64
}

// NewPhotoHandler creates a new photo handler
func NewPhotoHandler(photoService *photo.Service, maxUploadSize int64) *PhotoHandler {
	return &PhotoHandler{
		photoService:  photoService,
		maxUploadSize: maxUploadSize,
	}
}

// Upload uploads a new photo
// @Summary Upload photo
// @Description Upload a new photo with metadata
// @Tags Photos
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Photo file (JPG/PNG)"
// @Param raw_file formData file false "RAW file (optional)"
// @Param title formData string true "Photo title" maxLength(100)
// @Param description formData string false "Photo description" maxLength(500)
// @Param aircraft_type formData string false "Aircraft type"
// @Param airline formData string false "Airline"
// @Param registration formData string false "Aircraft registration"
// @Param airport formData string false "Airport (ICAO/IATA)"
// @Param category_id formData int false "Category ID"
// @Param tags formData string false "Tags (comma-separated)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/photos [post]
func (h *PhotoHandler) Upload(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get file
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file provided")
		return
	}

	// Early validation: check file size at handler level (fast fail)
	if file.Size > h.maxUploadSize {
		response.Error(c, http.StatusRequestEntityTooLarge, response.CodeValidationError, "File too large")
		return
	}

	// Get optional raw file
	rawFile, _ := c.FormFile("raw_file")

	// Get title (required)
	title := c.PostForm("title")
	if title == "" {
		response.BadRequest(c, "Title is required")
		return
	}
	if len(title) > 100 {
		response.BadRequest(c, "Title must be less than 100 characters")
		return
	}

	// Get other optional fields
	description := c.PostForm("description")
	if len(description) > 500 {
		response.BadRequest(c, "Description must be less than 500 characters")
		return
	}

	categoryID, _ := strconv.ParseInt(c.PostForm("category_id"), 10, 32)

	req := &photo.UploadRequest{
		UserID:       userID,
		File:         file,
		RawFile:      rawFile,
		Title:        title,
		Description:  description,
		AircraftType: c.PostForm("aircraft_type"),
		Airline:      c.PostForm("airline"),
		Registration: c.PostForm("registration"),
		Airport:      c.PostForm("airport"),
		CategoryID:   int32(categoryID),
		Tags:         c.PostForm("tags"),
	}

	result, err := h.photoService.Upload(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, storage.ErrFileTooLarge) {
			response.Error(c, http.StatusRequestEntityTooLarge, response.CodeValidationError, "File too large")
			return
		}
		if errors.Is(err, storage.ErrInvalidFileType) {
			response.BadRequest(c, "Invalid file type. Only JPG and PNG are allowed")
			return
		}
		response.InternalError(c, "Failed to upload photo")
		return
	}

	response.Success(c, result)
}

// List lists photos with pagination and filters
// @Summary List photos
// @Description Get a paginated list of approved photos with optional filters
// @Tags Photos
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param category_id query int false "Filter by category ID"
// @Param aircraft_type query string false "Filter by aircraft type"
// @Param airline query string false "Filter by airline"
// @Param airport query string false "Filter by airport"
// @Param keyword query string false "Search keyword"
// @Param sort_by query string false "Sort by: created_at, view_count, like_count, favorite_count"
// @Param sort_order query string false "Sort order: asc, desc"
// @Success 200 {object} response.Response
// @Router /api/v1/photos [get]
func (h *PhotoHandler) List(c *gin.Context) {
	var req photo.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	result, err := h.photoService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to list photos")
		return
	}

	response.Success(c, result)
}

// GetDetail gets photo detail by ID
// @Summary Get photo detail
// @Description Get detailed information about a photo
// @Tags Photos
// @Produce json
// @Param id path int true "Photo ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id} [get]
func (h *PhotoHandler) GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID")
		return
	}

	// Get current user ID if authenticated
	var currentUserID *int64
	if userID, exists := middleware.GetUserID(c); exists {
		currentUserID = &userID
	}

	detail, err := h.photoService.GetDetail(c.Request.Context(), id, currentUserID)
	if err != nil {
		if errors.Is(err, photo.ErrPhotoNotFound) {
			response.NotFound(c, "Photo not found")
			return
		}
		response.InternalError(c, "Failed to get photo detail")
		return
	}

	response.Success(c, detail)
}

// ListMine lists current user's photos
// @Summary List my photos
// @Description Get current user's uploaded photos
// @Tags Photos
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/photos/mine [get]
func (h *PhotoHandler) ListMine(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	result, err := h.photoService.ListMyPhotos(c.Request.Context(), userID, page, pageSize, status)
	if err != nil {
		response.InternalError(c, "Failed to list photos")
		return
	}

	response.Success(c, result)
}

// ListUserPhotos lists a user's approved photos
// @Summary List user's photos
// @Description Get a user's approved photos
// @Tags Photos
// @Produce json
// @Param id path int true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/photos [get]
func (h *PhotoHandler) ListUserPhotos(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.photoService.ListUserPhotos(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list photos")
		return
	}

	response.Success(c, result)
}

// ListFavorites lists current user's favorite photos
// @Summary List favorites
// @Description Get current user's favorite photos
// @Tags Photos
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/photos/favorites [get]
func (h *PhotoHandler) ListFavorites(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.photoService.ListFavorites(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list favorites")
		return
	}

	response.Success(c, result)
}

// AddFavorite adds a photo to favorites
// @Summary Add to favorites
// @Description Add a photo to current user's favorites
// @Tags Photos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id}/favorite [post]
func (h *PhotoHandler) AddFavorite(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID")
		return
	}

	err = h.photoService.AddFavorite(c.Request.Context(), userID, photoID)
	if err != nil {
		if errors.Is(err, photo.ErrPhotoNotFound) {
			response.NotFound(c, "Photo not found")
			return
		}
		response.InternalError(c, "Failed to add to favorites")
		return
	}

	response.Success(c, gin.H{"message": "Added to favorites"})
}

// RemoveFavorite removes a photo from favorites
// @Summary Remove from favorites
// @Description Remove a photo from current user's favorites
// @Tags Photos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id}/favorite [delete]
func (h *PhotoHandler) RemoveFavorite(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID")
		return
	}

	err = h.photoService.RemoveFavorite(c.Request.Context(), userID, photoID)
	if err != nil {
		if errors.Is(err, photo.ErrNotFavorited) {
			response.Error(c, http.StatusBadRequest, response.CodeValidationError, "Photo is not in favorites")
			return
		}
		response.InternalError(c, "Failed to remove from favorites")
		return
	}

	response.Success(c, gin.H{"message": "Removed from favorites"})
}

// AddLike adds a like to a photo
// @Summary Like photo
// @Description Add a like to a photo
// @Tags Photos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id}/like [post]
func (h *PhotoHandler) AddLike(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID")
		return
	}

	err = h.photoService.AddLike(c.Request.Context(), userID, photoID)
	if err != nil {
		if errors.Is(err, photo.ErrPhotoNotFound) {
			response.NotFound(c, "Photo not found")
			return
		}
		response.InternalError(c, "Failed to like photo")
		return
	}

	response.Success(c, gin.H{"message": "Liked"})
}

// RemoveLike removes a like from a photo
// @Summary Unlike photo
// @Description Remove a like from a photo
// @Tags Photos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id}/like [delete]
func (h *PhotoHandler) RemoveLike(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID")
		return
	}

	err = h.photoService.RemoveLike(c.Request.Context(), userID, photoID)
	if err != nil {
		if errors.Is(err, photo.ErrNotLiked) {
			response.Error(c, http.StatusBadRequest, response.CodeValidationError, "Photo is not liked")
			return
		}
		response.InternalError(c, "Failed to unlike photo")
		return
	}

	response.Success(c, gin.H{"message": "Unliked"})
}

// Delete deletes a photo
// @Summary Delete photo
// @Description Delete a photo (owner or admin only)
// @Tags Photos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/photos/{id} [delete]
func (h *PhotoHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID")
		return
	}

	// Check if user is admin
	roleStr, _ := middleware.GetRole(c)
	role := model.UserRole(roleStr)
	isAdmin := role == model.RoleAdmin || role == model.RoleSuperAdmin

	err = h.photoService.Delete(c.Request.Context(), photoID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, photo.ErrPhotoNotFound) {
			response.NotFound(c, "Photo not found")
			return
		}
		if errors.Is(err, photo.ErrNotOwner) {
			response.Forbidden(c, "You are not the owner of this photo")
			return
		}
		response.InternalError(c, "Failed to delete photo")
		return
	}

	response.Success(c, gin.H{"message": "Photo deleted"})
}
