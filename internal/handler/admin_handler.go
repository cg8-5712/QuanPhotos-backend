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

// ============================================
// Photo Review Handlers
// ============================================

// ListReviews lists photos pending review
// @Summary List pending reviews (Admin)
// @Description Get a paginated list of photos pending manual review
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status: ai_passed, ai_rejected, pending, all"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/reviews [get]
func (h *AdminHandler) ListReviews(c *gin.Context) {
	var req admin.ListReviewsRequest
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

	result, err := h.adminService.ListReviews(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to list reviews")
		return
	}

	response.Success(c, result)
}

// ReviewPhoto performs a manual review on a photo
// @Summary Review photo (Admin)
// @Description Approve or reject a photo
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Param request body admin.ReviewRequest true "Review action"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/reviews/{id} [post]
func (h *AdminHandler) ReviewPhoto(c *gin.Context) {
	reviewerID, exists := c.Get("userID")
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

	var req admin.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.adminService.ReviewPhoto(c.Request.Context(), photoID, reviewerID.(int64), &req)
	if err != nil {
		if errors.Is(err, admin.ErrPhotoNotFound) {
			response.NotFound(c, "Photo not found")
			return
		}
		response.InternalError(c, "Failed to review photo")
		return
	}

	response.Success(c, gin.H{"message": "Photo reviewed successfully"})
}

// ============================================
// Admin Delete Photo Handler
// ============================================

// AdminDeletePhoto deletes a photo with reason
// @Summary Delete photo (Admin)
// @Description Delete a photo with reason
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Param request body admin.DeletePhotoRequest true "Delete reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/photos/{id} [delete]
func (h *AdminHandler) AdminDeletePhoto(c *gin.Context) {
	adminID, exists := c.Get("userID")
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

	var req admin.DeletePhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.adminService.AdminDeletePhoto(c.Request.Context(), photoID, adminID.(int64), &req)
	if err != nil {
		if errors.Is(err, admin.ErrPhotoNotFound) {
			response.NotFound(c, "Photo not found")
			return
		}
		response.InternalError(c, "Failed to delete photo")
		return
	}

	response.Success(c, gin.H{"message": "Photo deleted successfully"})
}

// ============================================
// Ticket Management Handlers
// ============================================

// ListTickets lists all tickets for admin
// @Summary List tickets (Admin)
// @Description Get a paginated list of all tickets
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status: open, processing, resolved, closed"
// @Param type query string false "Filter by type: appeal, report, bug, feedback, other"
// @Param user_id query int false "Filter by user ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/tickets [get]
func (h *AdminHandler) ListTickets(c *gin.Context) {
	var req admin.AdminListTicketsRequest
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

	result, err := h.adminService.AdminListTickets(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to list tickets")
		return
	}

	response.Success(c, result)
}

// ProcessTicket processes a ticket
// @Summary Process ticket (Admin)
// @Description Update ticket status and optionally reply
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body admin.ProcessTicketRequest true "Process info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/tickets/{id} [put]
func (h *AdminHandler) ProcessTicket(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req admin.ProcessTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.adminService.ProcessTicket(c.Request.Context(), ticketID, adminID.(int64), &req)
	if err != nil {
		if errors.Is(err, admin.ErrTicketNotFound) {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalError(c, "Failed to process ticket")
		return
	}

	response.Success(c, result)
}

// ============================================
// Featured Photos Handlers
// ============================================

// AddFeatured adds a photo to featured list
// @Summary Add featured photo (Admin)
// @Description Add a photo to the featured list
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admin.AddFeaturedRequest true "Featured info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/admin/featured [post]
func (h *AdminHandler) AddFeatured(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req admin.AddFeaturedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.adminService.AddFeatured(c.Request.Context(), adminID.(int64), &req)
	if err != nil {
		if errors.Is(err, admin.ErrPhotoNotFound) {
			response.NotFound(c, "Photo not found")
			return
		}
		if errors.Is(err, admin.ErrAlreadyFeatured) {
			response.Conflict(c, "Photo is already featured")
			return
		}
		response.InternalError(c, "Failed to add featured photo")
		return
	}

	response.Success(c, result)
}

// RemoveFeatured removes a photo from featured list
// @Summary Remove featured photo (Admin)
// @Description Remove a photo from the featured list
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/featured/{id} [delete]
func (h *AdminHandler) RemoveFeatured(c *gin.Context) {
	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid photo ID")
		return
	}

	err = h.adminService.RemoveFeatured(c.Request.Context(), photoID)
	if err != nil {
		if errors.Is(err, admin.ErrNotFeatured) {
			response.NotFound(c, "Photo is not featured")
			return
		}
		response.InternalError(c, "Failed to remove featured photo")
		return
	}

	response.Success(c, gin.H{"message": "Photo removed from featured list"})
}

// ============================================
// Announcement Handlers
// ============================================

// ListAnnouncements lists announcements for admin
// @Summary List announcements (Admin)
// @Description Get a paginated list of announcements
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status: draft, published, all"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/announcements [get]
func (h *AdminHandler) ListAnnouncements(c *gin.Context) {
	var req admin.ListAnnouncementsRequest
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

	result, err := h.adminService.ListAnnouncements(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to list announcements")
		return
	}

	response.Success(c, result)
}

// CreateAnnouncement creates a new announcement
// @Summary Create announcement (Admin)
// @Description Create a new announcement
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admin.CreateAnnouncementRequest true "Announcement data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/announcements [post]
func (h *AdminHandler) CreateAnnouncement(c *gin.Context) {
	authorID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req admin.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.adminService.CreateAnnouncement(c.Request.Context(), authorID.(int64), &req)
	if err != nil {
		response.InternalError(c, "Failed to create announcement")
		return
	}

	response.Created(c, result)
}

// GetAnnouncement gets an announcement by ID
// @Summary Get announcement (Admin)
// @Description Get an announcement by ID
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Announcement ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/announcements/{id} [get]
func (h *AdminHandler) GetAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	result, err := h.adminService.GetAnnouncement(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, admin.ErrAnnouncementNotFound) {
			response.NotFound(c, "Announcement not found")
			return
		}
		response.InternalError(c, "Failed to get announcement")
		return
	}

	response.Success(c, result)
}

// UpdateAnnouncement updates an announcement
// @Summary Update announcement (Admin)
// @Description Update an announcement
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Announcement ID"
// @Param request body admin.UpdateAnnouncementRequest true "Announcement data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/announcements/{id} [put]
func (h *AdminHandler) UpdateAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	var req admin.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.adminService.UpdateAnnouncement(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, admin.ErrAnnouncementNotFound) {
			response.NotFound(c, "Announcement not found")
			return
		}
		response.InternalError(c, "Failed to update announcement")
		return
	}

	response.Success(c, result)
}

// DeleteAnnouncement deletes an announcement
// @Summary Delete announcement (Admin)
// @Description Delete an announcement
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Announcement ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/announcements/{id} [delete]
func (h *AdminHandler) DeleteAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	err = h.adminService.DeleteAnnouncement(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, admin.ErrAnnouncementNotFound) {
			response.NotFound(c, "Announcement not found")
			return
		}
		response.InternalError(c, "Failed to delete announcement")
		return
	}

	response.Success(c, gin.H{"message": "Announcement deleted successfully"})
}
