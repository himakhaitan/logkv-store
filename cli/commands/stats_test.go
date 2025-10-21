package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	servertypes "github.com/himakhaitan/logkv-store/types"
	"github.com/stretchr/testify/assert"
)

func TestStatsCommand_Success(t *testing.T) {
	resp := servertypes.StatsResponse{
		BaseResponse: servertypes.BaseResponse{
			Success: true,
		},
		TotalKeys: 5,
		TotalSize: 1234,
		Segments:  2,
	}
	data, _ := json.Marshal(resp)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/stats", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer server.Close()

	os.Setenv("LOGKV_ADDR", server.URL)
	defer os.Unsetenv("LOGKV_ADDR")
	cmd := NewStatsCommand()

	output := captureOutput(func() {
		cmd.SetArgs([]string{})
		_ = cmd.Execute()
	})
	assert.Contains(t, output, "[SUCCESS]", "Output should contain the SUCCESS tag.")
	assert.Contains(t, output, "Database Statistics", "Output should contain the main success message.")
	assert.Contains(t, output, "Total Keys: 5", "Output should contain the correct Total Keys.")
	assert.Contains(t, output, "Total Size: 1234 bytes", "Output should contain the correct Total Size (including 'bytes').") // Must include 'bytes'
	assert.Contains(t, output, "Segments: 2", "Output should contain the correct Segments count.")
}
func TestStatsCommand_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	os.Setenv("LOGKV_ADDR", server.URL)
	defer os.Unsetenv("LOGKV_ADDR")

	cmd := NewStatsCommand()
	output := captureOutput(func() {
		_ = cmd.Execute()
	})

	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "Server error")
}

func TestStatsCommand_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	os.Setenv("LOGKV_ADDR", server.URL)
	defer os.Unsetenv("LOGKV_ADDR")

	cmd := NewStatsCommand()
	output := captureOutput(func() {
		_ = cmd.Execute()
	})

	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "Invalid response")
}

func TestStatsCommand_UnsuccessfulResponse(t *testing.T) {
	resp := servertypes.StatsResponse{
		BaseResponse: servertypes.BaseResponse{
			Success: false,
			Message: "Database offline",
		},
	}
	data, _ := json.Marshal(resp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer server.Close()

	os.Setenv("LOGKV_ADDR", server.URL)
	defer os.Unsetenv("LOGKV_ADDR")

	cmd := NewStatsCommand()
	output := captureOutput(func() {
		_ = cmd.Execute()
	})

	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "Database offline")
}

func TestStatsCommand_ConnectionFailure(t *testing.T) {
	os.Setenv("LOGKV_ADDR", "http://127.0.0.1:1")
	defer os.Unsetenv("LOGKV_ADDR")

	cmd := NewStatsCommand()
	output := captureOutput(func() {
		_ = cmd.Execute()
	})

	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "Failed to connect to server")
}
