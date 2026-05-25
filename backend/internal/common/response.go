package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondOK[T any](c *gin.Context, status int, data T) {
	c.JSON(status, APIResponse[T]{
		Success: true,
		Data:    data,
	})
}

func RespondNoContent(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse[map[string]any]{
		Success: true,
		Data:    map[string]any{},
	})
}

func RespondError(c *gin.Context, err error) {
	appErr := NormalizeError(err)
	c.JSON(appErr.Status, APIResponse[any]{
		Success: false,
		Error: &APIError{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	})
}
