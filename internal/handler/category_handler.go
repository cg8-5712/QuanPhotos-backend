package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/category"
)

// CategoryHandler handles category HTTP requests
type CategoryHandler struct {
	categoryService *category.Service
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryService *category.Service) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// List lists all categories
// @Summary List categories
// @Description Get a list of all categories with photo counts
// @Tags Categories
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Success 200 {object} response.Response
// @Router /api/v1/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	var req category.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	result, err := h.categoryService.List(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to list categories")
		return
	}

	response.Success(c, result)
}

// GetByID gets a category by ID
// @Summary Get category
// @Description Get a category by ID
// @Tags Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/categories/{id} [get]
func (h *CategoryHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	result, err := h.categoryService.GetByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			response.NotFound(c, "Category not found")
			return
		}
		response.InternalError(c, "Failed to get category")
		return
	}

	response.Success(c, result)
}

// Create creates a new category (Admin only)
// @Summary Create category (Admin)
// @Description Create a new category
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body category.CreateRequest true "Category data"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req category.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.categoryService.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, category.ErrDuplicateName) {
			response.Conflict(c, "Category name already exists")
			return
		}
		response.InternalError(c, "Failed to create category")
		return
	}

	response.Created(c, result)
}

// Update updates a category (Admin only)
// @Summary Update category (Admin)
// @Description Update a category
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body category.UpdateRequest true "Category data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	var req category.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.categoryService.Update(c.Request.Context(), int32(id), &req)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			response.NotFound(c, "Category not found")
			return
		}
		if errors.Is(err, category.ErrDuplicateName) {
			response.Conflict(c, "Category name already exists")
			return
		}
		response.InternalError(c, "Failed to update category")
		return
	}

	response.Success(c, result)
}

// Delete deletes a category (Admin only)
// @Summary Delete category (Admin)
// @Description Delete a category
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param force query bool false "Force delete even if has photos"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	forceStr := c.Query("force")
	force := forceStr == "true"

	err = h.categoryService.Delete(c.Request.Context(), int32(id), force)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			response.NotFound(c, "Category not found")
			return
		}
		if errors.Is(err, category.ErrHasPhotos) {
			response.Error(c, http.StatusBadRequest, response.CodeValidationError, "Category has photos, use force=true to delete")
			return
		}
		response.InternalError(c, "Failed to delete category")
		return
	}

	response.Success(c, gin.H{"message": "Category deleted successfully"})
}
