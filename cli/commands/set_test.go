package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSetCommand_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v1/kv", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	os.Setenv("LOGKV_ADDR", server.URL)
	defer os.Unsetenv("LOGKV_ADDR")
	cmd := NewSetCommand()
	output := captureOutput(func() {
		cmd.SetArgs([]string{"foo", "bar"})
		_ = cmd.Execute()
	})
	assert.Contains(t, output, "[SUCCESS]")
	assert.Contains(t, output, "Set foo = bar")
}

func TestNewSetCommand_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	os.Setenv("LOGKV_ADDR", server.URL)
	defer os.Unsetenv("LOGKV_ADDR")
	cmd := NewSetCommand()
	output := captureOutput(func() {
		cmd.SetArgs([]string{"foo", "bar"})
		_ = cmd.Execute()
	})
	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "Server error")
}

func TestNewSetCommand_ConnectionFailure(t *testing.T) {
	os.Setenv("LOGKV_ADDR", "http://127.0.0.1:1")
	defer os.Unsetenv("LOGKV_ADDR")
	cmd := NewSetCommand()
	output := captureOutput(func() {
		cmd.SetArgs([]string{"foo", "bar"})
		_ = cmd.Execute()
	})
	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "Failed to connect to server")
}

func TestNewSetCommand_ArgValidation(t *testing.T) {
	cmd := NewSetCommand()
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s)")
}
