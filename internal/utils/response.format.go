package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorDetail `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorDetail provides detailed error information
type ErrorDetail struct {
	Code    string                 `json:"code,omitempty"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Items      interface{}     `json:"items"`
	Pagination *PaginationMeta `json:"pagination"`
}

// Success sends a successful JSON response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := APIResponse{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}

	c.JSON(statusCode, resp)
}

// Created sends a 201 Created response
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error JSON response
func Error(c *gin.Context, statusCode int, message string, err interface{}) {
	errorDetail := buildErrorDetail(err)
	
	resp := APIResponse{
		Status:    "error",
		Message:   message,
		Error:     errorDetail,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}

	c.JSON(statusCode, resp)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusBadRequest, message, err)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, nil)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, nil)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, nil)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusConflict, message, err)
}

// UnprocessableEntity sends a 422 Unprocessable Entity response
func UnprocessableEntity(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusUnprocessableEntity, message, err)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusInternalServerError, message, err)
}

// ValidationError sends a validation error response with field details
func ValidationError(c *gin.Context, errors map[string]string) {
	errorDetail := &ErrorDetail{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: make(map[string]interface{}),
	}

	for field, msg := range errors {
		errorDetail.Details[field] = msg
	}

	resp := APIResponse{
		Status:    "error",
		Message:   "Validation failed",
		Error:     errorDetail,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}

	c.JSON(http.StatusBadRequest, resp)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, message string, items interface{}, page, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	pagination := &PaginationMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		TotalPages:  totalPages,
	}

	data := &PaginatedResponse{
		Items:      items,
		Pagination: pagination,
	}

	Success(c, http.StatusOK, message, data)
}

// buildErrorDetail constructs error details from various error types
func buildErrorDetail(err interface{}) *ErrorDetail {
	if err == nil {
		return nil
	}

	errorDetail := &ErrorDetail{}

	switch e := err.(type) {
	case error:
		errorDetail.Message = e.Error()
	case string:
		errorDetail.Message = e
	case map[string]interface{}:
		if msg, ok := e["message"].(string); ok {
			errorDetail.Message = msg
		}
		if code, ok := e["code"].(string); ok {
			errorDetail.Code = code
		}
		errorDetail.Details = e
	case map[string]string:
		details := make(map[string]interface{})
		for k, v := range e {
			details[k] = v
		}
		errorDetail.Details = details
	default:
		errorDetail.Message = "An error occurred"
	}

	return errorDetail
}

// getRequestID extracts request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("requestID"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// CustomError allows creating custom error responses
type CustomError struct {
	Code       string
	Message    string
	Details    map[string]interface{}
	StatusCode int
}

// NewCustomError creates a new custom error
func NewCustomError(code, message string, statusCode int) *CustomError {
	return &CustomError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetail adds a detail field to the custom error
func (e *CustomError) WithDetail(key string, value interface{}) *CustomError {
	e.Details[key] = value
	return e
}

// Send sends the custom error response
func (e *CustomError) Send(c *gin.Context) {
	errorDetail := &ErrorDetail{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}

	resp := APIResponse{
		Status:    "error",
		Message:   e.Message,
		Error:     errorDetail,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}

	c.JSON(e.StatusCode, resp)
}