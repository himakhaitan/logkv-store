package cli

import (
	"github.com/himakhaitan/logkv-store/cli/commands"
	"github.com/spf13/cobra"
)

type CLI struct {
	root *cobra.Command
}

func NewCLI() *CLI {
	cli := &CLI{}

	rootCmd := &cobra.Command{
		Use:   "logkv-cli",
		Short: "A log-structured key-value store CLI",
		Long:  "LogKV CLI is a command-line interface for the LogKV key-value store",
	}

	// Create command registry and register all commands
	registry := commands.NewCommandRegistry()
	registry.RegisterCommands(rootCmd)

	cli.root = rootCmd

	return cli
}

func (c *CLI) Run() error {
	return c.root.Execute()
}
