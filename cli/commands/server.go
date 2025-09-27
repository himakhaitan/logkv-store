package commands

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/himakhaitan/logkv-store/cli/output"
	"github.com/spf13/cobra"
)

// NewServerCommand creates a new server command
func NewServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the LogKV server",
		Run: func(cmd *cobra.Command, args []string) {

			if runtime.GOOS != "windows" {
				output.Info(fmt.Sprintf("Server commands are not implemented for %v in CLI mode.", runtime.GOOS))
				output.Info("Use 'logkvd' binary to start the server.")
				return
			}

			output.Info("Starting LogKV server...")
			if err := buildServer(); err != nil {
				output.Error(fmt.Sprintf("Error in building the Server: %v", err))
				return
			}
			output.Info("Server Built successfully.")

			if err := startServer(); err != nil {
				output.Error(fmt.Sprintf("Error in starting the Server: %v", err))
				return
			}
			output.Info("Server started successfully")
		},
	}
}

func buildServer() error {
	cmd := exec.Command("go", "build", "-o", "./bin/logkvd.exe", "./cmd/logkvd/main.go")
	return cmd.Run()
}

func startServer() error {
	cmd := exec.Command("cmd", "/C", "start", "./bin/logkvd.exe")
	return cmd.Run()
}
