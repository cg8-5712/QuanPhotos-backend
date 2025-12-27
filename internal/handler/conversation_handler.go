package handler

import (
	"errors"
	"net/http"
	"strconv"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/conversation"

	"github.com/gin-gonic/gin"
)

// ConversationHandler handles conversation-related requests
type ConversationHandler struct {
	convService *conversation.Service
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(convService *conversation.Service) *ConversationHandler {
	return &ConversationHandler{
		convService: convService,
	}
}

// List godoc
// @Summary Get conversations
// @Description Get paginated list of conversations for current user
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=conversation.ListResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/conversations [get]
func (h *ConversationHandler) List(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req conversation.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.convService.List(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// Create godoc
// @Summary Create conversation and send message
// @Description Create a new conversation with a user and send the first message
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body conversation.CreateRequest true "Message data"
// @Success 201 {object} response.Response{data=conversation.CreateResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/conversations [post]
func (h *ConversationHandler) Create(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req conversation.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.convService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, conversation.ErrCannotMessageSelf) {
			response.BadRequest(c, "cannot send message to yourself")
			return
		}
		if errors.Is(err, conversation.ErrUserNotFound) {
			response.NotFound(c, "user not found")
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

// GetMessages godoc
// @Summary Get conversation messages
// @Description Get paginated messages in a conversation
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Success 200 {object} response.Response{data=conversation.GetMessagesResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/conversations/{id} [get]
func (h *ConversationHandler) GetMessages(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid conversation id")
		return
	}

	userID := c.GetInt64("userID")

	var req conversation.GetMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.convService.GetMessages(c.Request.Context(), userID, convID, &req)
	if err != nil {
		if errors.Is(err, conversation.ErrNotParticipant) {
			response.Forbidden(c, "you are not a participant of this conversation")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// SendMessage godoc
// @Summary Send message in conversation
// @Description Send a message in an existing conversation
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Param request body conversation.SendMessageRequest true "Message content"
// @Success 201 {object} response.Response{data=conversation.MessageItem}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/conversations/{id} [post]
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid conversation id")
		return
	}

	userID := c.GetInt64("userID")

	var req conversation.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.convService.SendMessage(c.Request.Context(), userID, convID, &req)
	if err != nil {
		if errors.Is(err, conversation.ErrNotParticipant) {
			response.Forbidden(c, "you are not a participant of this conversation")
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
// @Summary Delete conversation
// @Description Delete a conversation for current user
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Conversation ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/conversations/{id} [delete]
func (h *ConversationHandler) Delete(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid conversation id")
		return
	}

	userID := c.GetInt64("userID")

	err = h.convService.Delete(c.Request.Context(), userID, convID)
	if err != nil {
		if errors.Is(err, conversation.ErrNotParticipant) {
			response.Forbidden(c, "you are not a participant of this conversation")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
