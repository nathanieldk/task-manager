package response

import (
	"github.com/nathanieldk/task-manager/internal/dto"
	"github.com/nathanieldk/task-manager/internal/pkg/apperror"

	"github.com/labstack/echo/v4"
)

func Success(c echo.Context, statusCode int, data any) error {
	return c.JSON(statusCode, dto.NewSuccessResponse(data))
}

func Error(c echo.Context, err *apperror.AppError) error {
	return c.JSON(err.HTTPStatus, dto.NewErrorResponse(err.Code, err.Message))
}

func Paginated(c echo.Context, statusCode int, data any, pagination dto.PaginationMeta) error {
	return c.JSON(statusCode, dto.NewPaginatedResponse(data, pagination))
}
