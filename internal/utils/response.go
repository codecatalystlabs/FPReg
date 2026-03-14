package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type APIError struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: data})
}

func RespondMessage(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: msg})
}

func RespondPaginated(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data, Meta: &meta})
}

func RespondError(c *gin.Context, status int, msg string) {
	c.JSON(status, APIError{Success: false, Message: msg})
	c.Abort()
}

func RespondValidationError(c *gin.Context, errors []ErrorDetail) {
	c.JSON(http.StatusUnprocessableEntity, APIError{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
	c.Abort()
}

func RespondUnauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "Unauthorized"
	}
	RespondError(c, http.StatusUnauthorized, msg)
}

func RespondForbidden(c *gin.Context) {
	RespondError(c, http.StatusForbidden, "Forbidden: insufficient permissions")
}

func RespondNotFound(c *gin.Context, entity string) {
	RespondError(c, http.StatusNotFound, entity+" not found")
}
