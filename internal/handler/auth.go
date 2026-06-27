package handler

import (
	"net/http"

	"github.com/nathanieldk/task-manager/internal/dto"
	"github.com/nathanieldk/task-manager/internal/pkg/apperror"
	"github.com/nathanieldk/task-manager/internal/pkg/response"
	"github.com/nathanieldk/task-manager/internal/usecase"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// Register handles user registration.
// POST /auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperror.ErrBadRequest("INVALID_REQUEST", "Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return response.Error(c, apperror.ErrValidation(err.Error()))
	}

	result, err := h.authUsecase.Register(c.Request().Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	return response.Success(c, http.StatusCreated, result)
}

// Login handles user login.
// POST /auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperror.ErrBadRequest("INVALID_REQUEST", "Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return response.Error(c, apperror.ErrValidation(err.Error()))
	}

	result, err := h.authUsecase.Login(c.Request().Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr)
		}
		return response.Error(c, apperror.ErrInternal(err))
	}

	return response.Success(c, http.StatusOK, result)
}
