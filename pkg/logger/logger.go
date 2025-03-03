package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func SetupLogger() *zap.Logger {
	atomicLevel := zap.NewAtomicLevelAt(zap.InfoLevel)

	config := zap.Config{
		Level:            atomicLevel,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()

	if err != nil {
		fmt.Printf("Failed to create logger :%v\n", err)
		return nil
	}
	defer logger.Sync()

	if logger == nil {
		fmt.Println("Logger is nil!")
		return nil
	}
	logger.Info("Test info message")
	return logger
}
