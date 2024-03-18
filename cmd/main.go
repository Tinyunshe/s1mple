package main

import (
	"s1mple/pkg/config"
	"s1mple/pkg/logger"
	"s1mple/pkg/server"
)

func main() {
	config := config.NewConfig()

	logger := logger.NewLogger(config)
	defer logger.Sync()

	app := server.NewServer(config, logger)
	app.Run()
}
