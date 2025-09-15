package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/himakhaitan/logkv-store/cli/output"
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
				os.Exit(1)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				output.Error(fmt.Sprintf("Server error: %s\n", resp.Status))
				os.Exit(1)
			}
			b, _ := io.ReadAll(resp.Body)
			output.Info(fmt.Sprintf("Database Statistics:\n%s\n", string(b)))
		},
	}
}
