package main

import (
	"s1mple/config"
	"s1mple/server"

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

	app := server.Server{Config: config, Logger: logger}
	app.Run()
}
