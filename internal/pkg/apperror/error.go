package apperror

import (
	"fmt"
	"net/http"
)

type AppError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
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

func ErrBadRequest(code, message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       code,
		Message:    message,
	}
}

func ErrUnauthorized(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

func ErrForbidden(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
	}
}

func ErrNotFound(resource string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
	}
}

func ErrConflict(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
	}
}

func ErrValidation(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusUnprocessableEntity,
		Code:       "VALIDATION_ERROR",
		Message:    message,
	}
}

func ErrInternal(err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    "An internal server error occurred",
		Err:        err,
	}
}

func Wrap(err error, code string, message string, httpStatus int) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}
