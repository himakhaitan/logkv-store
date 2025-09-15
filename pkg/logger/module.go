package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module(service string) fx.Option {
	return fx.Provide(
		func() (*zap.Logger, error) {
			return New(service)
		},
	)
}
