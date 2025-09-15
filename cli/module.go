package cli

import "go.uber.org/fx"

var Module = fx.Provide(NewCLI)
