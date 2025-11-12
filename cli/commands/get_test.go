package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	servertypes "github.com/himakhaitan/logkv-store/types"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand(t *testing.T) {
	const testKey = "testkey"
	const testValue = "testvalue"
	const testTimestamp int64 = 1678886400

	successfulResponse := servertypes.GetResponse{
		Key:       testKey,
		Value:     testValue,
		Timestamp: testTimestamp,
	}
	successBody, _ := json.Marshal(successfulResponse)

	noTimestampResponse := servertypes.GetResponse{
		Key:       "notime",
		Value:     "val",
		Timestamp: 0,
	}
	noTimestampBody, _ := json.Marshal(noTimestampResponse)

	tests := []struct {
		name              string
		key               string
		handlerStatusCode int
		handlerBody       []byte
		expectedPath      string
	}{
		{
			name:              "Success_200_ValidJSON",
			key:               testKey,
			handlerStatusCode: http.StatusOK,
			handlerBody:       successBody,
			expectedPath:      "/v1/kv/testkey",
		},
		{
			name:              "Success_200_NoTimestamp",
			key:               "notime",
			handlerStatusCode: http.StatusOK,
			handlerBody:       noTimestampBody,
			expectedPath:      "/v1/kv/notime",
		},
		{
			name:              "KeyNotFound_404",
			key:               "missing",
			handlerStatusCode: http.StatusNotFound,
			handlerBody:       nil,
			expectedPath:      "/v1/kv/missing",
		},
		{
			name:              "ServerError_500",
			key:               "servererr",
			handlerStatusCode: http.StatusInternalServerError,
			handlerBody:       nil,
			expectedPath:      "/v1/kv/servererr",
		},
		{
			name:              "InvalidJSON_200",
			key:               testKey,
			handlerStatusCode: http.StatusOK,
			handlerBody:       []byte(`{"key": "testkey", "value": 123}`),
			expectedPath:      "/v1/kv/testkey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCapture := make(chan *http.Request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCapture <- r
				w.WriteHeader(tt.handlerStatusCode)
				if len(tt.handlerBody) > 0 {
					w.Write(tt.handlerBody)
				}
			}))
			defer server.Close()
			os.Setenv("LOGKV_ADDR", server.URL)
			defer os.Unsetenv("LOGKV_ADDR")
			cmd := NewGetCommand()
			executeCommand(t, cmd, []string{tt.key})
			select {
			case req := <-requestCapture:
				assert.Equal(t, http.MethodGet, req.Method, "Request method should be GET")
				assert.Equal(t, tt.expectedPath, req.URL.Path, "Request URL path should include the key")
			case <-time.After(50 * time.Millisecond):
				t.Fatalf("Timeout: Command did not make an HTTP request for key: %s", tt.key)
			}
		})
	}
}

func TestGetCommand_NetworkFailure(t *testing.T) {
	const failAddr = "http://127.0.0.1:1"
	os.Setenv("LOGKV_ADDR", failAddr)
	defer os.Unsetenv("LOGKV_ADDR")
	cmd := NewGetCommand()
	executeCommand(t, cmd, []string{"failkey"})
}

func TestGetCommand_ArgumentValidation(t *testing.T) {
	cmd := NewGetCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err, "Should require exactly one argument")
	cmd.SetArgs([]string{"k1", "k2"})
	err = cmd.Execute()
	assert.Error(t, err, "Should require exactly one argument")
}
