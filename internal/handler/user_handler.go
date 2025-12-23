package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/user"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	userService *user.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetCurrentUser gets current authenticated user's profile
// @Summary Get current user
// @Description Get current authenticated user's profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	profile, err := h.userService.GetProfile(c.Request.Context(), userID.(int64))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, "Failed to get user profile")
		return
	}

	response.Success(c, profile)
}

// UpdateCurrentUser updates current authenticated user's profile
// @Summary Update current user
// @Description Update current authenticated user's profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body user.UpdateProfileRequest true "Profile update info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req user.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, "Failed to update profile")
		return
	}

	response.Success(c, profile)
}

// GetUser gets a user's public profile by ID
// @Summary Get user by ID
// @Description Get a user's public profile by ID
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	info, err := h.userService.GetPublicInfo(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, "Failed to get user")
		return
	}

	response.Success(c, info)
}

// ChangePassword changes current user's password
// @Summary Change password
// @Description Change current authenticated user's password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body user.ChangePasswordRequest true "Password change info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req user.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		if errors.Is(err, user.ErrInvalidPassword) {
			response.Error(c, http.StatusBadRequest, response.CodeValidationError, "Current password is incorrect")
			return
		}
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, "Failed to change password")
		return
	}

	response.Success(c, gin.H{"message": "Password changed successfully"})
}
