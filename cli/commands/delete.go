package commands

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/himakhaitan/logkv-store/cli/output"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates a new delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			addr := os.Getenv("LOGKV_ADDR")
			if addr == "" {
				addr = "http://localhost:8080"
			}

			client := &http.Client{Timeout: 10 * time.Second}
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/kv/%s", addr, key), nil)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to create request: %v\n", err))
				os.Exit(1)
			}
			resp, err := client.Do(req)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to connect to server at %s\n %v", addr, err))
				os.Exit(1)
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				output.Info(fmt.Sprintf("Key '%s' not found\n", key))
				return
			}
			if resp.StatusCode != http.StatusNoContent {
				output.Error(fmt.Sprintf("Server error: %s\n", resp.Status))
				os.Exit(1)
			}
			output.Info(fmt.Sprintf("Deleted key: %s\n", key))
		},
	}
}
