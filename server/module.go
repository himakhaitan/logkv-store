package server

import (
	"github.com/himakhaitan/logkv-store/engine"
	"go.uber.org/fx"
)

// Module provides the HTTP server wired with fx
func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewMux),
		fx.Provide(NewHTTPServer),
		fx.Invoke(RegisterHooks),
		engine.Module(),
	)
}
