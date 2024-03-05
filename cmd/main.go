package main

import (
	"fmt"
	"os"
	"s1mple/config"
	serve "s1mple/server"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := serve.Server{Config: config}
	app.Run()
}
