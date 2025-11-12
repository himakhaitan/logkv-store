package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand(t *testing.T) {
	tests := []struct {
		name              string
		key               string
		handlerStatusCode int
		expectedPath      string
		expectedOutcome   string
	}{
		{
			name:              "Success_StatusNoContent",
			key:               "testkey1",
			handlerStatusCode: http.StatusNoContent,
			expectedPath:      "/v1/kv/testkey1",
			expectedOutcome:   "Success output should be called.",
		},
		{
			name:              "KeyNotFound_404",
			key:               "notfound",
			handlerStatusCode: http.StatusNotFound,
			expectedPath:      "/v1/kv/notfound",
			expectedOutcome:   "Warning output should be called.",
		},
		{
			name:              "ServerError_500",
			key:               "servererr",
			handlerStatusCode: http.StatusInternalServerError,
			expectedPath:      "/v1/kv/servererr",
			expectedOutcome:   "Error output should be called for server status.",
		},
		{
			name:              "ClientError_400",
			key:               "badreq",
			handlerStatusCode: http.StatusBadRequest,
			expectedPath:      "/v1/kv/badreq",
			expectedOutcome:   "Error output should be called for bad client request status.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This channel will capture the request details made to the server.
			requestCapture := make(chan *http.Request, 1)

			// Setup HTTP test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCapture <- r
				w.WriteHeader(tt.handlerStatusCode)
			}))
			defer server.Close()
			os.Setenv("LOGKV_ADDR", server.URL)
			defer os.Unsetenv("LOGKV_ADDR")
			cmd := NewDeleteCommand()
			executeCommand(t, cmd, []string{tt.key})
			select {
			case req := <-requestCapture:
				assert.Equal(t, http.MethodDelete, req.Method, "Request method should be DELETE")
				assert.Equal(t, tt.expectedPath, req.URL.Path, "Request URL path should include the key")
			case <-time.After(50 * time.Millisecond):
				t.Fatal("Timeout: Command did not make an HTTP request")
			}
		})
	}
}

func TestDeleteCommand_NetworkFailure(t *testing.T) {
	const failAddr = "http://127.0.0.1:1"
	os.Setenv("LOGKV_ADDR", failAddr)
	defer os.Unsetenv("LOGKV_ADDR")
	cmd := NewDeleteCommand()
	executeCommand(t, cmd, []string{"failkey"})
}

func TestDeleteCommand_ArgumentValidation(t *testing.T) {
	cmd := NewDeleteCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err, "Should require exactly one argument")
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0", "Cobra error message should confirm argument mismatch")
	cmd.SetArgs([]string{"k1", "k2"})
	err = cmd.Execute()
	assert.Error(t, err, "Should require exactly one argument")
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 2", "Cobra error message should confirm argument mismatch")
}
