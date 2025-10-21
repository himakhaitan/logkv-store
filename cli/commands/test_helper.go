package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// executeCommand runs the cobra command with given arguments.
// This helper is shared across all test files in the 'commands' package.
func executeCommand(t *testing.T, cmd *cobra.Command, args []string) {
	cmd.SetArgs(args)
	// We only check for cobra errors (arg count), not runtime errors (logged via output.Error)
	err := cmd.Execute()
	assert.NoError(t, err)
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)
	return buf.String()
}
