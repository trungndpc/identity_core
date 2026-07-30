package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

var (
	ErrNotFound       = New("NOT_FOUND", "resource not found", http.StatusNotFound)
	ErrBadRequest     = New("BAD_REQUEST", "invalid request", http.StatusBadRequest)
	ErrUnauthorized   = New("UNAUTHORIZED", "unauthorized", http.StatusUnauthorized)
	ErrForbidden      = New("FORBIDDEN", "forbidden", http.StatusForbidden)
	ErrConflict       = New("CONFLICT", "resource conflict", http.StatusConflict)
	ErrInternal       = New("INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
	ErrTenantRequired = New("TENANT_REQUIRED", "X-Tenant-Code header is required", http.StatusBadRequest)
	ErrTenantInvalid  = New("TENANT_INVALID", "tenant not found or inactive", http.StatusForbidden)
)

func Is(err, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == target.Code
	}
	return false
}
