package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes
const (
	CodeSuccess           = 0
	CodeInvalidParams     = 40001
	CodeValidationFailed  = 40002
	CodeValidationError   = 40002 // Alias for CodeValidationFailed
	CodeUnauthorized      = 40101
	CodeTokenExpired      = 40102
	CodeTokenInvalid      = 40103
	CodeForbidden         = 40301
	CodeNotFound          = 40401
	CodeConflict          = 40901
	CodeUnsupportedFormat = 42201
	CodeFileTooLarge      = 42202
	CodeTooManyRequests   = 42901
	CodeInternalError     = 50001
	CodeDatabaseError     = 50002
	CodeAIServiceError    = 50003
)

// Error messages
var errorMessages = map[int]string{
	CodeSuccess:           "success",
	CodeInvalidParams:     "invalid parameters",
	CodeValidationFailed:  "validation failed",
	CodeUnauthorized:      "unauthorized",
	CodeTokenExpired:      "token expired",
	CodeTokenInvalid:      "invalid token",
	CodeForbidden:         "forbidden",
	CodeNotFound:          "resource not found",
	CodeConflict:          "resource conflict",
	CodeUnsupportedFormat: "unsupported file format",
	CodeFileTooLarge:      "file too large",
	CodeTooManyRequests:   "too many requests",
	CodeInternalError:     "internal server error",
	CodeDatabaseError:     "database error",
	CodeAIServiceError:    "AI service unavailable",
}

// Response represents the standard API response format
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedData represents paginated data
type PagedData struct {
	List       interface{} `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Success sends a success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage sends a success response with custom message
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// SuccessWithPagination sends a success response with pagination
func SuccessWithPagination(c *gin.Context, list interface{}, page, pageSize, total int) {
	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: PagedData{
			List: list,
			Pagination: Pagination{
				Page:       page,
				PageSize:   pageSize,
				Total:      total,
				TotalPages: totalPages,
			},
		},
	})
}

// Error sends an error response with HTTP status, code, and message
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithCode sends an error response based on error code
func ErrorWithCode(c *gin.Context, code int) {
	message := errorMessages[code]
	if message == "" {
		message = "unknown error"
	}
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithMessage sends an error response with custom message
func ErrorWithMessage(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithData sends an error response with additional data
func ErrorWithData(c *gin.Context, code int, data interface{}) {
	message := errorMessages[code]
	if message == "" {
		message = "unknown error"
	}
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// getHTTPStatus maps error code to HTTP status code
func getHTTPStatus(code int) int {
	switch {
	case code == CodeSuccess:
		return http.StatusOK
	case code >= 40001 && code < 40100:
		return http.StatusBadRequest
	case code >= 40101 && code < 40200:
		return http.StatusUnauthorized
	case code >= 40301 && code < 40400:
		return http.StatusForbidden
	case code >= 40401 && code < 40500:
		return http.StatusNotFound
	case code >= 40901 && code < 41000:
		return http.StatusConflict
	case code >= 42201 && code < 42300:
		return http.StatusUnprocessableEntity
	case code >= 42901 && code < 43000:
		return http.StatusTooManyRequests
	case code >= 50001:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Abort sends an error response and aborts the request
func Abort(c *gin.Context, code int) {
	ErrorWithCode(c, code)
	c.Abort()
}

// AbortWithMessage sends an error response with message and aborts
func AbortWithMessage(c *gin.Context, code int, message string) {
	ErrorWithMessage(c, code, message)
	c.Abort()
}

// Convenience functions for common responses

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeInvalidParams,
		Message: message,
	})
}

// ValidationError sends a 400 response for validation errors
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeValidationFailed,
		Message: message,
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
}

// TokenExpired sends a 401 response for expired tokens
func TokenExpired(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeTokenExpired,
		Message: message,
	})
}

// TokenInvalid sends a 401 response for invalid tokens
func TokenInvalid(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeTokenInvalid,
		Message: message,
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    CodeConflict,
		Message: message,
	})
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Code:    CodeTooManyRequests,
		Message: message,
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: message,
	})
}
