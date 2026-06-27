package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathanieldk/task-manager/config"
	configapp "github.com/nathanieldk/task-manager/config/app"
	configlogger "github.com/nathanieldk/task-manager/config/logger"
	configpostgres "github.com/nathanieldk/task-manager/config/postgres"
	configredis "github.com/nathanieldk/task-manager/config/redis"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Zap logger
	logger, err := configlogger.Init(cfg.Logger.FilePath)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Connect to PostgreSQL
	db, err := configpostgres.Connect(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL")

	// Run database migrations
	if err := configpostgres.RunMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}
	logger.Info("Database migrations completed")

	// Connect to Redis
	ctx := context.Background()
	rdb, err := configredis.Connect(ctx, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer rdb.Close()
	logger.Info("Connected to Redis")

	// Initialize application layers (repos, usecases, handlers)
	appInstance := configapp.Initialize(db, rdb, cfg, logger)

	// Initialize Echo router
	e := configapp.NewRouter(cfg, logger, appInstance)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	go func() {
		logger.Info("Starting server", zap.String("address", addr))
		if err := e.Start(addr); err != nil {
			logger.Info("Server stopped", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	if err := e.Close(); err != nil {
		logger.Fatal("Failed to shut down server", zap.Error(err))
	}
	logger.Info("Server shut down gracefully")
}
