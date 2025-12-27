package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/middleware"
	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/ticket"
)

// TicketHandler handles ticket HTTP requests
type TicketHandler struct {
	ticketService *ticket.Service
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketService *ticket.Service) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// Create creates a new ticket
// @Summary Create ticket
// @Description Create a new ticket (appeal, report, or other)
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ticket.CreateRequest true "Ticket data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/tickets [post]
func (h *TicketHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req ticket.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.ticketService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, "Failed to create ticket")
		return
	}

	response.SuccessWithMessage(c, "ticket created", result)
}

// List lists user's tickets with pagination
// @Summary List my tickets
// @Description Get a paginated list of current user's tickets
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status: open, processing, resolved, closed"
// @Param type query string false "Filter by type: appeal, report, other"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/tickets [get]
func (h *TicketHandler) List(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req ticket.ListRequest
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

	result, err := h.ticketService.List(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, "Failed to list tickets")
		return
	}

	response.Success(c, result)
}

// GetDetail gets ticket detail by ID
// @Summary Get ticket detail
// @Description Get detailed information about a ticket including replies
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/tickets/{id} [get]
func (h *TicketHandler) GetDetail(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	detail, err := h.ticketService.GetDetail(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ticket.ErrTicketNotFound) {
			response.NotFound(c, "Ticket not found")
			return
		}
		if errors.Is(err, ticket.ErrNotOwner) {
			response.Forbidden(c, "You don't have permission to view this ticket")
			return
		}
		response.InternalError(c, "Failed to get ticket detail")
		return
	}

	response.Success(c, detail)
}

// Reply adds a reply to a ticket
// @Summary Reply to ticket
// @Description Add a reply to an existing ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param body body ticket.ReplyRequest true "Reply content"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/tickets/{id}/replies [post]
func (h *TicketHandler) Reply(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
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

	var req ticket.ReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.ticketService.Reply(c.Request.Context(), userID, ticketID, &req)
	if err != nil {
		if errors.Is(err, ticket.ErrTicketNotFound) {
			response.NotFound(c, "Ticket not found")
			return
		}
		if errors.Is(err, ticket.ErrNotOwner) {
			response.Forbidden(c, "You don't have permission to reply to this ticket")
			return
		}
		if errors.Is(err, ticket.ErrTicketClosed) {
			response.BadRequest(c, "Cannot reply to a closed ticket")
			return
		}
		response.InternalError(c, "Failed to submit reply")
		return
	}

	response.SuccessWithMessage(c, "reply submitted", result)
}
