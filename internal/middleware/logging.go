package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			latency := time.Since(start)
			status := c.Response().Status
			requestID, _ := c.Get("request_id").(string)

			fields := []zap.Field{
				zap.String("request_id", requestID),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.String("query", c.Request().URL.RawQuery),
				zap.Int("status", status),
				zap.Duration("latency", latency),
				zap.String("ip", c.RealIP()),
				zap.String("user_agent", c.Request().UserAgent()),
			}

			switch {
			case status >= 500:
				logger.Error("Server error", fields...)
			case status >= 400:
				logger.Warn("Client error", fields...)
			default:
				logger.Info("Request", fields...)
			}

			return nil
		}
	}
}
