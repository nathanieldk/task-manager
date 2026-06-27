package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes a new zap.Logger that writes to both stdout and the specified file.
func Init(filePath string) (*zap.Logger, error) {
	if filePath != "" {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Readable timestamps

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	var cores []zapcore.Core

	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.InfoLevel))

	if filePath != "" {
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(file), zap.InfoLevel))
	}

	combinedCore := zapcore.NewTee(cores...)
	log := zap.New(combinedCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return log, nil
}
