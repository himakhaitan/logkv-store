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

// NewStatsCommand creates a new stats command
func NewStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show database statistics",
		Run: func(cmd *cobra.Command, args []string) {
			addr := os.Getenv("LOGKV_ADDR")
			if addr == "" {
				addr = "http://localhost:8080"
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(fmt.Sprintf("%s/v1/stats", addr))
			if err != nil {
				output.Error(fmt.Sprintf("Failed to connect to server at %s\n %v", addr, err))
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				output.Error(fmt.Sprintf("Server error: %s", resp.Status))
				return
			}
			var out servertypes.StatsResponse
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
			output.Success("Database Statistics")
			output.Info(fmt.Sprintf("Total Keys: %d", out.TotalKeys))
			output.Info(fmt.Sprintf("Total Size: %d bytes", out.TotalSize))
			output.Info(fmt.Sprintf("Segments: %d", out.Segments))
		},
	}
}
