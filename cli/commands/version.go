package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "v0.1.0"

// NewVersionCommand creates a new version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("logkv-cli version %s\n", Version)
		},
	}
}
