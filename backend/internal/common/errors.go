package common

import (
	"errors"
	"net/http"
	"strings"
)

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return e.Message
}

func NewError(status int, code string, message string) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

func BadRequest(message string) *AppError {
	return NewError(http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func Unauthorized(message string) *AppError {
	return NewError(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(message string) *AppError {
	return NewError(http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(message string) *AppError {
	return NewError(http.StatusNotFound, "NOT_FOUND", message)
}

func Conflict(message string) *AppError {
	return NewError(http.StatusConflict, "CONFLICT", message)
}

func TooManyRequests(message string) *AppError {
	return NewError(http.StatusTooManyRequests, "RATE_LIMITED", message)
}

func Internal(message string) *AppError {
	return NewError(http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func NormalizeError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return Internal("Unexpected server error")
}

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
