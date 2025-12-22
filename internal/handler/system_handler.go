package handler

import (
	"net/http"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/system"

	"github.com/gin-gonic/gin"
)

// SystemHandler handles system-related requests
type SystemHandler struct {
	service *system.Service
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(service *system.Service) *SystemHandler {
	return &SystemHandler{
		service: service,
	}
}

// Health handles health check requests
// GET /health
func (h *SystemHandler) Health(c *gin.Context) {
	status, healthy := h.service.GetHealth()

	httpStatus := http.StatusOK
	if !healthy {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, status)
}

// Info handles system info requests
// GET /api/v1/system/info
func (h *SystemHandler) Info(c *gin.Context) {
	info := h.service.GetSystemInfo()
	response.Success(c, info)
}
