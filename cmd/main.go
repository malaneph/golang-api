package main

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2/humacli"
	"golang-api/api"
	"golang-api/config"
)

func main() {
	appConfig, err := config.New()
	if err != nil {
		slog.Error("config error: %v", err)
		panic(err)
	}

	cli := humacli.New(func(hooks humacli.Hooks, c *config.AppConfig) {
		// Create a new router & API
		mux := api.SetupRouter()

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			slog.Info("Starting server on port", "port", appConfig.Server.Port)
			err := http.ListenAndServe(":"+appConfig.Server.Port, mux)
			if err != nil {
				slog.Error("Server failed", "error", err)
				panic(err)
			}
		})
	})

	cli.Run()
}
