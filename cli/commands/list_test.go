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

func TestListCommand(t *testing.T) {
	tests := []struct {
		name              string
		handlerStatusCode int
		handlerResponse   servertypes.ListKeysResponse
		expectedPath      string
		bodyIsInvalid     bool
	}{
		{
			name:              "Success_WithKeys",
			handlerStatusCode: http.StatusOK,
			handlerResponse: servertypes.ListKeysResponse{
				BaseResponse: servertypes.BaseResponse{Success: true},
				Keys:         []string{"k1", "k2", "k3"},
			},
			expectedPath: "/v1/keys",
		},
		{
			name:              "Success_NoKeys",
			handlerStatusCode: http.StatusOK,
			handlerResponse: servertypes.ListKeysResponse{
				BaseResponse: servertypes.BaseResponse{Success: true},
				Keys:         []string{},
			},
			expectedPath: "/v1/keys",
		},
		{
			name:              "Server_Reported_Failure_WithMessage",
			handlerStatusCode: http.StatusOK,
			handlerResponse: servertypes.ListKeysResponse{
				BaseResponse: servertypes.BaseResponse{Success: false, Message: "Database offline"},
				Keys:         nil,
			},
			expectedPath: "/v1/keys",
		},
		{
			name:              "ServerError_500",
			handlerStatusCode: http.StatusInternalServerError,
			handlerResponse:   servertypes.ListKeysResponse{},
			expectedPath:      "/v1/keys",
		},
		{
			name:              "InvalidJSON_DecodeError",
			handlerStatusCode: http.StatusOK,
			handlerResponse:   servertypes.ListKeysResponse{},
			expectedPath:      "/v1/keys",
			bodyIsInvalid:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCapture := make(chan *http.Request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCapture <- r
				w.WriteHeader(tt.handlerStatusCode)
				var body []byte
				if tt.bodyIsInvalid {
					body = []byte(`{"success": "maybe"}`)
				} else {
					body, _ = json.Marshal(tt.handlerResponse)
				}
				w.Write(body)
			}))
			defer server.Close()

			os.Setenv("LOGKV_ADDR", server.URL)
			defer os.Unsetenv("LOGKV_ADDR")

			cmd := NewListCommand()
			executeCommand(t, cmd, []string{})

			select {
			case req := <-requestCapture:
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Equal(t, tt.expectedPath, req.URL.Path)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Timeout: Command did not make an HTTP request")
			}
		})
	}
}

func TestListCommand_NetworkFailure(t *testing.T) {
	os.Setenv("LOGKV_ADDR", "http://127.0.0.1:1")
	defer os.Unsetenv("LOGKV_ADDR")
	cmd := NewListCommand()
	executeCommand(t, cmd, []string{})
}

func TestListCommand_ArgumentValidation(t *testing.T) {
	cmd0 := NewListCommand()
	executeCommand(t, cmd0, []string{})
	cmd1 := NewListCommand()
	cmd1.SetArgs([]string{"extra"})
	err := cmd1.Execute()
	assert.NoError(t, err, "List command ignores extra args unless Args validation is added")
}
