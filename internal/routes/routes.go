package routes

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/nathanieldk/task-manager/config"
	"github.com/nathanieldk/task-manager/internal/handler"
	"github.com/nathanieldk/task-manager/internal/middleware"
)

func Setup(
	e *echo.Echo,
	cfg *config.Config,
	logger *zap.Logger,
	authHandler *handler.AuthHandler,
	taskHandler *handler.TaskHandler,
) {
	e.Use(middleware.Recovery(logger))
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(logger))
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization, "Idempotency-Key"},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	auth := e.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	tasks := e.Group("/tasks", middleware.JWTAuth(cfg.JWT.Secret))
	tasks.POST("", taskHandler.Create)
	tasks.GET("", taskHandler.List)
	tasks.GET("/:id", taskHandler.GetByID)
	tasks.PUT("/:id", taskHandler.Update)
	tasks.DELETE("/:id", taskHandler.Delete)
	tasks.POST("/:id/assign", taskHandler.Assign)
}
