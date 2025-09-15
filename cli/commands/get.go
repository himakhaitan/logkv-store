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

// NewGetCommand creates a new get command
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a value by key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			addr := os.Getenv("LOGKV_ADDR")
			if addr == "" {
				addr = "http://localhost:8080"
			}

			client := &http.Client{Timeout: 10 * time.Second}
			url := fmt.Sprintf("%s/v1/kv/%s", addr, key)
			resp, err := client.Get(url)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to connect to server at %s\n %v", addr, err))
				os.Exit(1)
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				output.Info(fmt.Sprintf("Key '%s' not found\n", key))
				return
			}
			if resp.StatusCode != http.StatusOK {
				output.Error(fmt.Sprintf("Server error: %s\n", resp.Status))
				os.Exit(1)
			}
			var out struct {
				Value string `json:"value"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				output.Error(fmt.Sprintf("Invalid response: %v\n", err))
				os.Exit(1)
			}
			output.Info(fmt.Sprintf("%s\n", out.Value))
		},
	}
}
