package httperr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	CodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	CodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	CodeForbidden        ErrorCode = "FORBIDDEN"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeConflict         ErrorCode = "CONFLICT"
	CodeInternalError    ErrorCode = "INTERNAL_ERROR"
)

type errorBody struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id,omitempty"`
}

// ErrorResponse is the standard error envelope returned by all API endpoints.
type ErrorResponse struct {
	Error errorBody `json:"error"`
}

func requestID(c *gin.Context) string {
	id, _ := c.Get("request_id")
	s, _ := id.(string)
	return s
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: errorBody{Code: CodeValidationFailed, Message: message, RequestID: requestID(c)},
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error: errorBody{Code: CodeUnauthorized, Message: message, RequestID: requestID(c)},
	})
}

// AbortUnauthorized aborts the request chain and responds with 401.
func AbortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
		Error: errorBody{Code: CodeUnauthorized, Message: message, RequestID: requestID(c)},
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Error: errorBody{Code: CodeForbidden, Message: message, RequestID: requestID(c)},
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: errorBody{Code: CodeNotFound, Message: message, RequestID: requestID(c)},
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{
		Error: errorBody{Code: CodeConflict, Message: message, RequestID: requestID(c)},
	})
}

// Internal responds with 500 and a generic message. The real error must be logged by the caller before calling this.
func Internal(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: errorBody{Code: CodeInternalError, Message: "an internal error occurred", RequestID: requestID(c)},
	})
}
