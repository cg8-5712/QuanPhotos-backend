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

// Error sends an error response
func Error(c *gin.Context, code int) {
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
	Error(c, code)
	c.Abort()
}

// AbortWithMessage sends an error response with message and aborts
func AbortWithMessage(c *gin.Context, code int, message string) {
	ErrorWithMessage(c, code, message)
	c.Abort()
}
