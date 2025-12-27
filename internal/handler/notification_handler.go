package handler

import (
	"errors"
	"strconv"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/notification"

	"github.com/gin-gonic/gin"
)

// NotificationHandler handles notification-related requests
type NotificationHandler struct {
	notifService *notification.Service
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notifService *notification.Service) *NotificationHandler {
	return &NotificationHandler{
		notifService: notifService,
	}
}

// List godoc
// @Summary Get notifications
// @Description Get paginated list of notifications for current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type" Enums(like, comment, reply, follow, share, featured, review, system, message)
// @Param is_read query bool false "Filter by read status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=notification.ListResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req notification.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.notifService.List(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Get the count of unread notifications for current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=notification.UnreadCountResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/notifications/unread [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetInt64("userID")

	result, err := h.notifService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notifID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	userID := c.GetInt64("userID")

	err = h.notifService.MarkAsRead(c.Request.Context(), userID, notifID)
	if err != nil {
		if errors.Is(err, notification.ErrNotificationNotFound) {
			response.NotFound(c, "notification not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetInt64("userID")

	err := h.notifService.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
