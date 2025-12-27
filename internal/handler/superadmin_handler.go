package handler

import (
	"errors"
	"strconv"

	"QuanPhotos/internal/middleware"
	"QuanPhotos/internal/pkg/response"
	superadminService "QuanPhotos/internal/service/superadmin"

	"github.com/gin-gonic/gin"
)

// SuperadminHandler handles superadmin-related HTTP requests
type SuperadminHandler struct {
	service *superadminService.Service
}

// NewSuperadminHandler creates a new superadmin handler
func NewSuperadminHandler(service *superadminService.Service) *SuperadminHandler {
	return &SuperadminHandler{service: service}
}

// ListAdmins handles GET /api/v1/superadmin/admins
func (h *SuperadminHandler) ListAdmins(c *gin.Context) {
	result, err := h.service.ListAdmins(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetAdminPermissions handles GET /api/v1/superadmin/admins/:id/permissions
func (h *SuperadminHandler) GetAdminPermissions(c *gin.Context) {
	adminID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	permissions, err := h.service.GetAdminPermissions(c.Request.Context(), adminID)
	if err != nil {
		if errors.Is(err, superadminService.ErrNotAdmin) {
			response.NotFound(c, "admin not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"admin_id":    adminID,
		"permissions": permissions,
	})
}

// GrantPermissionsRequest represents the request to grant permissions
type GrantPermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required,min=1"`
}

// GrantPermissions handles POST /api/v1/superadmin/admins/:id/permissions
func (h *SuperadminHandler) GrantPermissions(c *gin.Context) {
	adminID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	var req GrantPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	grantedBy, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "authentication required")
		return
	}

	err = h.service.GrantPermissions(c.Request.Context(), adminID, req.Permissions, grantedBy)
	if err != nil {
		switch {
		case errors.Is(err, superadminService.ErrNotAdmin):
			response.NotFound(c, "admin not found")
		case errors.Is(err, superadminService.ErrInvalidPermission):
			response.BadRequest(c, "invalid permission")
		case errors.Is(err, superadminService.ErrSelfModify):
			response.Forbidden(c, "cannot modify own permissions")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"message": "permissions granted"})
}

// RevokePermissionsRequest represents the request to revoke permissions
type RevokePermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required,min=1"`
}

// RevokePermissions handles DELETE /api/v1/superadmin/admins/:id/permissions
func (h *SuperadminHandler) RevokePermissions(c *gin.Context) {
	adminID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid admin id")
		return
	}

	var req RevokePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	revokedBy, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "authentication required")
		return
	}

	err = h.service.RevokePermissions(c.Request.Context(), adminID, req.Permissions, revokedBy)
	if err != nil {
		switch {
		case errors.Is(err, superadminService.ErrNotAdmin):
			response.NotFound(c, "admin not found")
		case errors.Is(err, superadminService.ErrSelfModify):
			response.Forbidden(c, "cannot modify own permissions")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"message": "permissions revoked"})
}

// ListAvailablePermissions handles GET /api/v1/superadmin/permissions
func (h *SuperadminHandler) ListAvailablePermissions(c *gin.Context) {
	permissions := h.service.GetAllPermissions()
	response.Success(c, gin.H{
		"permissions": permissions,
	})
}

// ListReviewers handles GET /api/v1/superadmin/reviewers
func (h *SuperadminHandler) ListReviewers(c *gin.Context) {
	result, err := h.service.ListReviewers(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetReviewerCategories handles GET /api/v1/superadmin/reviewers/:id/categories
func (h *SuperadminHandler) GetReviewerCategories(c *gin.Context) {
	reviewerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid reviewer id")
		return
	}

	categoryIDs, err := h.service.GetReviewerCategories(c.Request.Context(), reviewerID)
	if err != nil {
		if errors.Is(err, superadminService.ErrNotReviewer) {
			response.NotFound(c, "reviewer not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"reviewer_id":  reviewerID,
		"category_ids": categoryIDs,
	})
}

// AssignCategoriesRequest represents the request to assign categories
type AssignCategoriesRequest struct {
	CategoryIDs []int `json:"category_ids" binding:"required,min=1"`
}

// AssignCategories handles POST /api/v1/superadmin/reviewers/:id/categories
func (h *SuperadminHandler) AssignCategories(c *gin.Context) {
	reviewerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid reviewer id")
		return
	}

	var req AssignCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.service.AssignCategories(c.Request.Context(), reviewerID, req.CategoryIDs)
	if err != nil {
		switch {
		case errors.Is(err, superadminService.ErrNotReviewer):
			response.NotFound(c, "reviewer not found")
		case errors.Is(err, superadminService.ErrCategoryNotFound):
			response.BadRequest(c, "category not found")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"message": "categories assigned"})
}

// RevokeCategoriesRequest represents the request to revoke categories
type RevokeCategoriesRequest struct {
	CategoryIDs []int `json:"category_ids" binding:"required,min=1"`
}

// RevokeCategories handles DELETE /api/v1/superadmin/reviewers/:id/categories
func (h *SuperadminHandler) RevokeCategories(c *gin.Context) {
	reviewerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid reviewer id")
		return
	}

	var req RevokeCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.service.RevokeCategories(c.Request.Context(), reviewerID, req.CategoryIDs)
	if err != nil {
		if errors.Is(err, superadminService.ErrNotReviewer) {
			response.NotFound(c, "reviewer not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "categories revoked"})
}

// GetUserRestrictions handles GET /api/v1/superadmin/users/:id/restrictions
func (h *SuperadminHandler) GetUserRestrictions(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	restrictions, err := h.service.GetUserRestrictions(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, superadminService.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, restrictions)
}

// UpdateUserRestrictionsRequest represents the request to update user restrictions
type UpdateUserRestrictionsRequest struct {
	CanComment *bool `json:"can_comment"`
	CanMessage *bool `json:"can_message"`
	CanUpload  *bool `json:"can_upload"`
}

// UpdateUserRestrictions handles PUT /api/v1/superadmin/users/:id/restrictions
func (h *SuperadminHandler) UpdateUserRestrictions(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req UpdateUserRestrictionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if at least one field is provided
	if req.CanComment == nil && req.CanMessage == nil && req.CanUpload == nil {
		response.BadRequest(c, "at least one restriction field must be provided")
		return
	}

	restrictions, err := h.service.UpdateUserRestrictions(c.Request.Context(), userID, req.CanComment, req.CanMessage, req.CanUpload)
	if err != nil {
		if errors.Is(err, superadminService.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, restrictions)
}
