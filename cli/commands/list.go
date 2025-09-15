package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/himakhaitan/logkv-store/cli/output"
	"github.com/spf13/cobra"
)

// NewListCommand creates a new list command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all keys",
		Run: func(cmd *cobra.Command, args []string) {
			addr := os.Getenv("LOGKV_ADDR")
			if addr == "" {
				addr = "http://localhost:8080"
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(fmt.Sprintf("%s/v1/keys", addr))
			if err != nil {
				output.Error(fmt.Sprintf("Failed to connect to server at %s\n %v", addr, err))
				os.Exit(1)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				output.Error(fmt.Sprintf("Server error: %s\n", resp.Status))
				os.Exit(1)
			}
			var keys []string
			if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
				output.Error(fmt.Sprintf("Invalid response: %v\n", err))
				os.Exit(1)
			}
			if len(keys) == 0 {
				output.Info("No keys found")
			} else {
				for _, key := range keys {
					output.Info(key)
				}
			}
		},
	}
}
