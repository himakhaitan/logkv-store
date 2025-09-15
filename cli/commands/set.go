package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/himakhaitan/logkv-store/cli/output"
	"github.com/spf13/cobra"
)

// NewSetCommand creates a new set command
func NewSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a key-value pair",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]
			addr := os.Getenv("LOGKV_ADDR")
			if addr == "" {
				addr = "http://localhost:8080"
			}

			client := &http.Client{Timeout: 10 * time.Second}
			body, _ := json.Marshal(map[string]string{"key": key, "value": value})
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/kv", addr), bytes.NewReader(body))
			if err != nil {
				output.Error(fmt.Sprintf("Failed to create request: %v", err))
				os.Exit(1)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to connect to server at %s\n %v", addr, err))
				os.Exit(1)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				output.Error(fmt.Sprintf("Server error: %s\n", resp.Status))
				os.Exit(1)
			}
			output.Success(fmt.Sprintf("Set %s = %s\n", key, value))
		},
	}
}
