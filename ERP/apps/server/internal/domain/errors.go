package domain

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

func NewNotFound(resource string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: fmt.Sprintf("%s not found", resource)}
}

func NewBadRequest(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg}
}

func NewForbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg}
}

func NewConflict(msg string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg}
}

func NewInternal(msg string) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg}
}

func IsNotFound(err error) bool {
	var e *AppError
	return errors.As(err, &e) && e.Code == http.StatusNotFound
}

// Sentinel errors for common cases
var (
	ErrInvalidPIN       = NewUnauthorized("invalid PIN")
	ErrTokenExpired     = NewUnauthorized("token expired")
	ErrTokenInvalid     = NewUnauthorized("invalid token")
	ErrInsufficientPerm = NewForbidden("insufficient permissions")
	ErrStopList         = NewBadRequest("product is on stop-list (insufficient stock)")
	ErrNegativeStock    = &AppError{Code: http.StatusConflict, Message: "operation results in negative stock", Detail: "recorded with needs_review flag"}
	ErrDuplicateUUID    = NewConflict("client_uuid already processed (duplicate event)")
)
