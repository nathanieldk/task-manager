package middleware

import (
	"net/http"
	"strings"

	"github.com/nathanieldk/task-manager/internal/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTAuth validates JWT tokens and injects user claims into the Echo context.
func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized,
					dto.NewErrorResponse("UNAUTHORIZED", "Missing authorization header"))
			}

			// Expect "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return c.JSON(http.StatusUnauthorized,
					dto.NewErrorResponse("UNAUTHORIZED", "Invalid authorization header format"))
			}

			tokenString := parts[1]

			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized,
					dto.NewErrorResponse("UNAUTHORIZED", "Invalid or expired token"))
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized,
					dto.NewErrorResponse("UNAUTHORIZED", "Invalid token claims"))
			}

			userID, _ := claims["user_id"].(string)
			teamIDFloat, _ := claims["team_id"].(float64)
			teamID := int(teamIDFloat)

			if userID == "" {
				return c.JSON(http.StatusUnauthorized,
					dto.NewErrorResponse("UNAUTHORIZED", "Invalid token: missing user_id"))
			}

			// Set into context
			c.Set("user_id", userID)
			c.Set("team_id", teamID)

			return next(c)
		}
	}
}
