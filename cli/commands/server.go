package commands

import (
	"github.com/himakhaitan/logkv-store/cli/output"
	"github.com/spf13/cobra"
)

// NewServerCommand creates a new server command
func NewServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the LogKV server",
		Run: func(cmd *cobra.Command, args []string) {
			output.Info("Starting LogKV server...")
			output.Info("Server commands are not implemented in CLI mode.")
			output.Info("Use 'logkvd' binary to start the server.")
		},
	}
}
