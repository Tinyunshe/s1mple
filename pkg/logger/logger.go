package logger

import (
	"s1mple/pkg/config"

	"go.uber.org/zap"
)

func NewLogger(config *config.Config) *zap.Logger {
	switch config.LogLevel {
	case "info":
		logger, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		return logger
	case "debug":
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		return logger
	}
	panic("No exisit this log level")
}
