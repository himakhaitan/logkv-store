package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/himakhaitan/logkv-store/cli/output"
	servertypes "github.com/himakhaitan/logkv-store/types"
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
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				output.Error(fmt.Sprintf("Server error: %s", resp.Status))
				return
			}
			var out servertypes.ListKeysResponse
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				output.Error(fmt.Sprintf("Invalid response: %v", err))
				return
			}
			if !out.Success {
				if out.Message != "" {
					output.Error(out.Message)
				} else {
					output.Error("Request failed")
				}
				return
			}
			if len(out.Keys) == 0 {
				output.Info("No keys found")
			} else {
				output.Success("Keys:")
				for _, key := range out.Keys {
					output.Info(key)
				}
			}
		},
	}
}
