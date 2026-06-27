package app

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/nathanieldk/task-manager/config"
	"github.com/nathanieldk/task-manager/internal/pkg/validator"
	"github.com/nathanieldk/task-manager/internal/routes"
)

// NewRouter initializes an Echo router and sets up global middlewares and routes.
func NewRouter(cfg *config.Config, logger *zap.Logger, appInstance *App) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.NewCustomValidator()

	// Setup routes and global middleware
	routes.Setup(e, cfg, logger, appInstance.AuthHandler, appInstance.TaskHandler)

	return e
}
