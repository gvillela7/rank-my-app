package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponseDoc represents a successful API response for Swagger documentation
type SuccessResponseDoc struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message" example:"Operation successful"`
}

// ErrorResponseDoc represents an error API response for Swagger documentation
type ErrorResponseDoc struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Detailed error message"`
	Message string `json:"message" example:"User-friendly error message"`
}

func SuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, err error, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   err.Error(),
		Message: message,
	})
}

func ValidationErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error:   err.Error(),
		Message: "Validation failed",
	})
}
