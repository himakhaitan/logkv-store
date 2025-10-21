package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewCommandRegistry(t *testing.T) {
	reg := NewCommandRegistry()
	assert.NotNil(t, reg)
}

func TestGetAllCommands(t *testing.T) {
	reg := NewCommandRegistry()
	commands := reg.GetAllCommands()
	assert.NotNil(t, commands)
	assert.NotEmpty(t, commands)
	assert.Len(t, commands, 7, "Expected 7 commands to be registered")
	expected := []string{"version", "get", "set", "delete", "list", "stats", "server"}
	for i, cmd := range commands {
		assert.Equal(t, expected[i], cmd.Name())
	}
}

func TestRegisterCommands(t *testing.T) {
	reg := NewCommandRegistry()
	rootCmd := &cobra.Command{Use: "logkv"}
	reg.RegisterCommands(rootCmd)
	subCmds := rootCmd.Commands()
	assert.Len(t, subCmds, 7)
	names := []string{}
	for _, c := range subCmds {
		names = append(names, c.Name())
	}
	expected := []string{"version", "get", "set", "delete", "list", "stats", "server"}
	assert.ElementsMatch(t, expected, names)
}
