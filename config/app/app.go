package app

import (
	"database/sql"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/nathanieldk/task-manager/config"
	"github.com/nathanieldk/task-manager/internal/handler"
	"github.com/nathanieldk/task-manager/internal/repository"
	"github.com/nathanieldk/task-manager/internal/usecase"
)

// App containers the instantiated HTTP handlers for the application.
type App struct {
	AuthHandler *handler.AuthHandler
	TaskHandler *handler.TaskHandler
}

// Initialize repositories, usecases, and handlers layers.
func Initialize(db *sql.DB, rdb *goredis.Client, cfg *config.Config, logger *zap.Logger) *App {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	taskLogRepo := repository.NewTaskLogRepository(db)
	txManager := repository.NewTransactionManager(db)

	// Initialize usecases
	idempotencyUsecase := usecase.NewIdempotencyUsecase(rdb)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg.JWT.Secret, cfg.JWT.ExpirationHours)
	taskUsecase := usecase.NewTaskUsecase(taskRepo, taskLogRepo, userRepo, txManager, idempotencyUsecase, logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	taskHandler := handler.NewTaskHandler(taskUsecase)

	return &App{
		AuthHandler: authHandler,
		TaskHandler: taskHandler,
	}
}
