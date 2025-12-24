package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/admin"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	adminService *admin.Service
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService *admin.Service) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// ListUsers lists all users with pagination and filters
// @Summary List users (Admin)
// @Description Get a paginated list of users with optional filters
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param keyword query string false "Search by username or email"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req admin.ListUsersRequest
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

	result, err := h.adminService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to list users")
		return
	}

	response.Success(c, result)
}

// UpdateUserRole updates a user's role
// @Summary Update user role (Admin)
// @Description Update a user's role
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body admin.UpdateRoleRequest true "Role update info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	// Get operator info
	operatorID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	operatorRoleStr, _ := c.Get("role")
	operatorRole := model.UserRole(operatorRoleStr.(string))

	// Get target user ID
	idStr := c.Param("id")
	targetUserID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req admin.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.adminService.UpdateUserRole(c.Request.Context(), operatorID.(int64), targetUserID, operatorRole, &req)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, admin.ErrCannotChangeSelf) {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "Cannot change your own role")
			return
		}
		if errors.Is(err, admin.ErrInsufficientPerm) {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "Insufficient permissions")
			return
		}
		if errors.Is(err, admin.ErrInvalidRole) {
			response.BadRequest(c, "Invalid role")
			return
		}
		response.InternalError(c, "Failed to update user role")
		return
	}

	response.Success(c, gin.H{"message": "Role updated"})
}

// UpdateUserStatus updates a user's status (ban/unban)
// @Summary Update user status (Admin)
// @Description Ban or unban a user
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body admin.UpdateStatusRequest true "Status update info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/users/{id}/status [put]
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	// Get operator info
	operatorID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	operatorRoleStr, _ := c.Get("role")
	operatorRole := model.UserRole(operatorRoleStr.(string))

	// Get target user ID
	idStr := c.Param("id")
	targetUserID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req admin.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.adminService.UpdateUserStatus(c.Request.Context(), operatorID.(int64), targetUserID, operatorRole, &req)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, admin.ErrCannotChangeSelf) {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "Cannot change your own status")
			return
		}
		if errors.Is(err, admin.ErrInsufficientPerm) {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "Insufficient permissions")
			return
		}
		response.InternalError(c, "Failed to update user status")
		return
	}

	response.Success(c, gin.H{"message": "Operation successful"})
}
