package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()

	// Assert the CLI structure is initialized
	assert.NotNil(t, cli, "NewCLI should return a non-nil CLI struct")
	assert.NotNil(t, cli.root, "CLI root command should be initialized")

	// Assert the root command attributes are correct
	assert.Equal(t, "logkv-cli", cli.root.Use, "Root command Use name is incorrect")
	assert.Contains(t, cli.root.Short, "log-structured key-value store CLI", "Root command Short description is incorrect")

	// Assert that subcommands have been registered
	subcommands := cli.root.Commands()
	assert.True(t, len(subcommands) > 0, "RegisterCommands should have added subcommands to the root command")

	hasSubcommand := false
	for _, cmd := range subcommands {
		if cmd.Use == "list" {
			hasSubcommand = true
			break
		}
	}
	assert.True(t, hasSubcommand, "The 'list' subcommand should be registered on the root command")
}

func TestCLIRun(t *testing.T) {
	cli := NewCLI()
	oldStdout := cli.root.OutOrStdout()
	oldStderr := cli.root.ErrOrStderr()
	defer func() {
		cli.root.SetOut(oldStdout)
		cli.root.SetErr(oldStderr)
	}()

	// Temporarily capture output
	output := bytes.NewBuffer(nil)
	cli.root.SetOut(output)
	cli.root.SetErr(output)

	// Test: Run with no arguments
	cli.root.SetArgs([]string{})
	err := cli.Run()

	// Assert no execution error occurred
	assert.NoError(t, err, "Running the CLI with no arguments should not return a core error")

	// Capture the printed output
	capturedOutput := output.String()

	// Assert that the Usage/Use name was printed
	assert.True(t, strings.Contains(capturedOutput, cli.root.Use), "Running with no args should print the usage message based on 'Use'")

	// Assert that the Long description was printed
	assert.True(t, strings.Contains(capturedOutput, cli.root.Long), "Running with no args should print the Long description in the help output")

	// Check for the 'Available Commands' header which proves subcommands were registered
	assert.True(t, strings.Contains(capturedOutput, "Available Commands:"), "Help output should list available commands.")
}
