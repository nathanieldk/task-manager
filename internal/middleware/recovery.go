package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/nathanieldk/task-manager/internal/dto"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from panics, logs the stack trace.
func Recovery(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Capture stack trace
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					stackTrace := string(buf[:n])

					requestID, _ := c.Get("request_id").(string)

					// Log panic error
					logger.Error("Panic recovered",
						zap.String("request_id", requestID),
						zap.String("method", c.Request().Method),
						zap.String("path", c.Request().URL.Path),
						zap.String("error", fmt.Sprintf("%v", r)),
						zap.String("stack", stackTrace),
					)

					// Default error
					c.JSON(http.StatusInternalServerError,
						dto.NewErrorResponse("INTERNAL_ERROR", "An internal server error occurred"))
				}
			}()

			return next(c)
		}
	}
}
