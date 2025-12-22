package middleware

import (
	"net/http"

	"QuanPhotos/internal/pkg/logger"
	"QuanPhotos/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a gin middleware for recovering from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code:    response.CodeInternalError,
					Message: "internal server error",
				})
			}
		}()
		c.Next()
	}
}
