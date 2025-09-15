package commands

import (
	"github.com/spf13/cobra"
)

// CommandRegistry holds all available commands
type CommandRegistry struct {
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{}
}

// GetAllCommands returns all available commands
func (r *CommandRegistry) GetAllCommands() []*cobra.Command {
	return []*cobra.Command{
		NewVersionCommand(),
		NewGetCommand(),
		NewSetCommand(),
		NewDeleteCommand(),
		NewListCommand(),
		NewStatsCommand(),
		NewServerCommand(),
	}
}

// RegisterCommands adds all commands to the root command
func (r *CommandRegistry) RegisterCommands(rootCmd *cobra.Command) {
	commands := r.GetAllCommands()
	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}
}
