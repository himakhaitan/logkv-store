package engine

import (
	"github.com/himakhaitan/logkv-store/store"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		store.Module,
		fx.Provide(NewDB),
	)
}
