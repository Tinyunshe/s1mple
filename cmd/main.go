package main

import (
	"s1mple/pkg/config"
	"s1mple/pkg/server"

	"go.uber.org/zap"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	app := server.NewServer(config, logger)
	app.Run()
}
