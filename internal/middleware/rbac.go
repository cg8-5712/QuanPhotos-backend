package middleware

import (
	"github.com/gin-gonic/gin"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/response"
)

// RequireRole creates middleware that requires specific role
func RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, exists := GetRole(c)
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		userRole := model.UserRole(roleStr)

		// Check if user has any of the required roles
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// RequireMinRole creates middleware that requires minimum role level
func RequireMinRole(minRole model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, exists := GetRole(c)
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		userRole := model.UserRole(roleStr)

		if model.RoleLevel(userRole) < model.RoleLevel(minRole) {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates middleware that requires admin or superadmin role
func RequireAdmin() gin.HandlerFunc {
	return RequireMinRole(model.RoleAdmin)
}

// RequireSuperAdmin creates middleware that requires superadmin role
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleSuperAdmin)
}

// RequireReviewer creates middleware that requires reviewer or higher role
func RequireReviewer() gin.HandlerFunc {
	return RequireMinRole(model.RoleReviewer)
}

// RequireUser creates middleware that requires user or higher role (not guest)
func RequireUser() gin.HandlerFunc {
	return RequireMinRole(model.RoleUser)
}

// RequireActiveUser creates middleware that requires active user
// This should be used with a user service to check user status
func RequireActiveUser(checkStatus func(c *gin.Context, userID int64) (bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		isActive, err := checkStatus(c, userID)
		if err != nil {
			response.InternalError(c, "Failed to check user status")
			c.Abort()
			return
		}

		if !isActive {
			response.Forbidden(c, "Your account is not active")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission creates middleware that checks if user can perform action
func RequirePermission(action string, checkPermission func(c *gin.Context, userID int64, action string) (bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		allowed, err := checkPermission(c, userID, action)
		if err != nil {
			response.InternalError(c, "Failed to check permission")
			c.Abort()
			return
		}

		if !allowed {
			response.Forbidden(c, "You don't have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}
