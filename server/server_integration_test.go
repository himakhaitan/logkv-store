package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/himakhaitan/logkv-store/engine"
	"github.com/himakhaitan/logkv-store/pkg/config"
	"github.com/himakhaitan/logkv-store/store"
	"github.com/himakhaitan/logkv-store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupIntegrationServer(t *testing.T) (*httptest.Server, *store.Store, *engine.DB, func()) {
	logger := zaptest.NewLogger(t)
	tmpDir, err := os.MkdirTemp("", "logkv_integration")
	require.NoError(t, err)

	cfg := &config.Config{DataDir: tmpDir}

	s, err := store.New(logger, cfg)
	require.NoError(t, err)

	db := &engine.DB{Store: s}
	mux := NewMux(db, logger)
	ts := httptest.NewServer(mux)

	cleanup := func() {
		ts.Close()
		s.Close()
		os.RemoveAll(tmpDir)
	}

	return ts, s, db, cleanup
}

func TestServerIntegration_AllEndpoints(t *testing.T) {
	ts, s, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// Health
	resp, err := http.Get(ts.URL + "/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// PUT /v1/kv
	setBody := `{"key":"foo","value":"bar"}`
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/v1/kv", bytes.NewBufferString(setBody))
	req.Header.Set("Content-Type", "application/json")
	resp2, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusNoContent, resp2.StatusCode)

	// GET /v1/kv/foo
	getResp, _ := http.Get(ts.URL + "/v1/kv/foo")
	var getData types.GetResponse
	_ = json.NewDecoder(getResp.Body).Decode(&getData)
	assert.Equal(t, "foo", getData.Key)
	assert.Equal(t, "bar", getData.Value)

	// DELETE /v1/kv/foo
	delReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/kv/foo", nil)
	delResp, _ := http.DefaultClient.Do(delReq)
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	// GET deleted key
	getResp2, _ := http.Get(ts.URL + "/v1/kv/foo")
	assert.Equal(t, http.StatusNotFound, getResp2.StatusCode)

	// List /v1/keys
	s.Set("a", "1")
	s.Set("b", "2")
	listResp, _ := http.Get(ts.URL + "/v1/keys")
	var listData types.ListKeysResponse
	_ = json.NewDecoder(listResp.Body).Decode(&listData)
	assert.ElementsMatch(t, []string{"a", "b"}, listData.Keys)

	// Stats /v1/stats
	statsRespRaw, _ := http.Get(ts.URL + "/v1/stats")
	var statsData types.StatsResponse
	_ = json.NewDecoder(statsRespRaw.Body).Decode(&statsData)
	assert.Equal(t, 2, statsData.TotalKeys)
	var totalSize int64
	for _, k := range []string{"a", "b"} {
		val, _ := s.Get(k)
		totalSize += int64(len(val))
	}
	assert.Equal(t, totalSize, statsData.TotalSize)
}

func TestServerIntegration_EmptyKeyAndMethodNotAllowed(t *testing.T) {
	ts, _, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// PUT with empty key
	emptyBody := `{"key":"","value":"val"}`
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/v1/kv", bytes.NewBufferString(emptyBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// POST with empty key
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/kv", bytes.NewBufferString(emptyBody))
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := http.DefaultClient.Do(req2)
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	// PUT with wrong method
	req3, _ := http.NewRequest(http.MethodPatch, ts.URL+"/v1/kv/foo", nil)
	resp3, _ := http.DefaultClient.Do(req3)
	assert.Equal(t, http.StatusMethodNotAllowed, resp3.StatusCode)

	// GET /v1/kv with wrong method
	req4, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/kv", nil)
	resp4, _ := http.DefaultClient.Do(req4)
	assert.Equal(t, http.StatusMethodNotAllowed, resp4.StatusCode)

	// GET /v1/keys with wrong method
	req5, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/keys", nil)
	resp5, _ := http.DefaultClient.Do(req5)
	assert.Equal(t, http.StatusMethodNotAllowed, resp5.StatusCode)

	// GET /v1/stats with wrong method
	req6, _ := http.NewRequest(http.MethodPut, ts.URL+"/v1/stats", nil)
	resp6, _ := http.DefaultClient.Do(req6)
	assert.Equal(t, http.StatusMethodNotAllowed, resp6.StatusCode)
}

func TestServerIntegration_InvalidJSON(t *testing.T) {
	ts, _, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	invalidJSON := `{"key":"foo","value":"bar"`
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/v1/kv", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestServerIntegration_DeleteNonExistentKey(t *testing.T) {
	ts, _, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/kv/nonexistent", nil)
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
