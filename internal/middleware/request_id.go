package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const RequestIDHeader = "X-Request-ID"

// RequestID generates a UUID request_id for each request and injects it into
// the context and response headers.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if client already sent a request ID
			reqID := c.Request().Header.Get(RequestIDHeader)
			if reqID == "" {
				reqID = uuid.New().String()
			}

			// Store in context for downstream use
			c.Set("request_id", reqID)

			// Add to response headers
			c.Response().Header().Set(RequestIDHeader, reqID)

			return next(c)
		}
	}
}
