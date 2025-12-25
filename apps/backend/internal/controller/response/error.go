package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes
const (
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeValidationError     = "VALIDATION_ERROR"
)

// ErrorBody represents the error structure in response
type ErrorBody struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// ErrorResponse represents the standard error response
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ValidationDetail represents a single validation error
type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorBody represents validation error structure
type ValidationErrorBody struct {
	Code    string             `json:"code,omitempty"`
	Message string             `json:"message"`
	Details []ValidationDetail `json:"details,omitempty"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error ValidationErrorBody `json:"error"`
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, details interface{}) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorBody{
			Code:    CodeBadRequest,
			Message: message,
		},
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error: ErrorBody{
			Code:    CodeUnauthorized,
			Message: message,
		},
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Error: ErrorBody{
			Code:    CodeForbidden,
			Message: message,
		},
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: ErrorBody{
			Code:    CodeNotFound,
			Message: message,
		},
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{
		Error: ErrorBody{
			Code:    CodeConflict,
			Message: message,
		},
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: ErrorBody{
			Code:    CodeInternalServerError,
			Message: "internal server error",
		},
	})
}

// ValidationError sends a 400 Bad Request with validation details
func ValidationError(c *gin.Context, details []ValidationDetail) {
	c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Error: ValidationErrorBody{
			Code:    CodeValidationError,
			Message: "validation failed",
			Details: details,
		},
	})
}
