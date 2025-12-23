package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/pkg/response"
	"QuanPhotos/internal/service/auth"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Registration info"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, tokens, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrUsernameExists) {
			response.Error(c, http.StatusConflict, response.CodeValidationError, "Username already exists")
			return
		}
		if errors.Is(err, auth.ErrEmailExists) {
			response.Error(c, http.StatusConflict, response.CodeValidationError, "Email already exists")
			return
		}
		response.InternalError(c, "Failed to register user")
		return
	}

	response.Created(c, gin.H{
		"user":   user.ToProfile(),
		"tokens": tokens,
	})
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, tokens, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid username or password")
			return
		}
		if errors.Is(err, auth.ErrUserBanned) {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "Your account has been banned")
			return
		}
		response.InternalError(c, "Failed to login")
		return
	}

	response.Success(c, gin.H{
		"user":   user.ToProfile(),
		"tokens": tokens,
	})
}

// Refresh handles token refresh
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req auth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrTokenExpired) {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenInvalid, "Invalid or expired refresh token")
			return
		}
		if errors.Is(err, auth.ErrUserNotActive) {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "Your account is not active")
			return
		}
		response.InternalError(c, "Failed to refresh token")
		return
	}

	response.Success(c, gin.H{
		"tokens": tokens,
	})
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshRequest true "Refresh token to invalidate"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req auth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Even without token, we consider logout successful
		response.Success(c, nil)
		return
	}

	_ = h.authService.Logout(c.Request.Context(), req.RefreshToken)
	response.Success(c, nil)
}
