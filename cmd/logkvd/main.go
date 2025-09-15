package main

import (
	"github.com/himakhaitan/logkv-store/pkg/config"
	"github.com/himakhaitan/logkv-store/pkg/logger"
	"github.com/himakhaitan/logkv-store/server"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module("logkv-server"),
		config.Module(),
		server.Module(),
	)

	app.Run()
}
