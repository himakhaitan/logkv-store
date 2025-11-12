package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/himakhaitan/logkv-store/server"
	"github.com/himakhaitan/logkv-store/store"
	"github.com/himakhaitan/logkv-store/types"

	"github.com/stretchr/testify/assert"
	fxt "go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

type mockDB struct {
	data   map[string]string
	stats  store.Stats
	errGet error
	errSet error
	errDel error
	errLst error
	errSta error
}

func (m *mockDB) Get(key string) (string, error) {
	if m.errGet != nil {
		return "", m.errGet
	}
	v, ok := m.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return v, nil
}

func (m *mockDB) Set(key, value string) error {
	if m.errSet != nil {
		return m.errSet
	}
	m.data[key] = value
	return nil
}

func (m *mockDB) Delete(key string) error {
	if m.errDel != nil {
		return m.errDel
	}
	if _, ok := m.data[key]; !ok {
		return errors.New("key not found")
	}
	delete(m.data, key)
	return nil
}

func (m *mockDB) List() ([]string, error) {
	if m.errLst != nil {
		return nil, m.errLst
	}
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockDB) Stats() (store.Stats, error) {
	if m.errSta != nil {
		return store.Stats{}, m.errSta
	}
	return m.stats, nil
}

// Helper Handlers

func makeKVHandler(db *mockDB, logger *zap.Logger) http.HandlerFunc {
	_ = logger
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		key := r.URL.Path[len("/v1/kv/"):]
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "missing key",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		switch r.Method {
		case http.MethodGet:
			value, err := db.Get(key)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(types.BaseResponse{
					Success:   false,
					Message:   err.Error(),
					Timestamp: time.Now().Unix(),
				})
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(types.GetResponse{
				Key:   key,
				Value: value,
				BaseResponse: types.BaseResponse{
					Success:   true,
					Timestamp: time.Now().Unix(),
					Message:   "key fetched successfully",
				},
			})
		case http.MethodDelete:
			if err := db.Delete(key); err != nil {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(types.BaseResponse{
					Success:   false,
					Message:   err.Error(),
					Timestamp: time.Now().Unix(),
				})
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "method not allowed",
				Timestamp: time.Now().Unix(),
			})
		}
	}
}

func makeSetHandler(db *mockDB, logger *zap.Logger) http.HandlerFunc {
	_ = logger
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "Method not allowed",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		var req types.SetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "invalid json",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		if req.Key == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "missing key",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		if err := db.Set(req.Key, req.Value); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   err.Error(),
				Timestamp: time.Now().Unix(),
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func makeListHandler(db *mockDB, logger *zap.Logger) http.HandlerFunc {
	_ = logger
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "Method not allowed",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		keys, err := db.List()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "Internal Server Error",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(types.ListKeysResponse{
			Keys: keys,
			BaseResponse: types.BaseResponse{
				Success:   true,
				Timestamp: time.Now().Unix(),
				Message:   "keys fetched successfully",
			},
		})
	}
}

func makeStatsHandler(db *mockDB, logger *zap.Logger) http.HandlerFunc {
	_ = logger
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "Method not allowed",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		stats, err := db.Stats()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{
				Success:   false,
				Message:   "Internal Server Error",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(types.StatsResponse{
			TotalKeys: stats.TotalKeys,
			TotalSize: stats.TotalSize,
			Segments:  stats.Segments,
			BaseResponse: types.BaseResponse{
				Success:   true,
				Timestamp: time.Now().Unix(),
				Message:   "stats fetched successfully",
			},
		})
	}
}

func setupTestServer() (*httptest.Server, *mockDB) {
	mock := &mockDB{data: make(map[string]string)}
	logger, _ := zap.NewDevelopment()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/kv/", makeKVHandler(mock, logger))
	mux.HandleFunc("/v1/kv", makeSetHandler(mock, logger))
	mux.HandleFunc("/v1/keys", makeListHandler(mock, logger))
	mux.HandleFunc("/v1/stats", makeStatsHandler(mock, logger))

	server := httptest.NewServer(mux)
	return server, mock
}

func TestHealthEndpoint(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetAndGetKV(t *testing.T) {
	server, db := setupTestServer()
	defer server.Close()

	body := `{"key":"foo","value":"bar"}`
	resp, err := http.Post(server.URL+"/v1/kv", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	v, ok := db.data["foo"]
	assert.True(t, ok)
	assert.Equal(t, "bar", v)

	getResp, err := http.Get(server.URL + "/v1/kv/foo")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var getBody types.GetResponse
	assert.NoError(t, json.NewDecoder(getResp.Body).Decode(&getBody))
	assert.Equal(t, "bar", getBody.Value)
}

func TestSetKV_MissingKey_PostPut(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		body := `{"key":"","value":"bar"}`
		req, _ := http.NewRequest(method, server.URL+"/v1/kv", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var res types.BaseResponse
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.False(t, res.Success)
		assert.Equal(t, "missing key", res.Message)
	}
}

func TestSetKV_InvalidJSON(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	body := `{"key":"foo",`
	resp, err := http.Post(server.URL+"/v1/kv", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var res types.BaseResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.False(t, res.Success)
	assert.Equal(t, "invalid json", res.Message)
}

func TestSetKV_DBError(t *testing.T) {
	server, db := setupTestServer()
	defer server.Close()
	db.errSet = errors.New("set fail")

	body := `{"key":"foo","value":"bar"}`
	resp, err := http.Post(server.URL+"/v1/kv", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var res types.BaseResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.False(t, res.Success)
	assert.Equal(t, "set fail", res.Message)
}

func TestGetKV_KeyNotFound(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/kv/missing")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteKV(t *testing.T) {
	server, db := setupTestServer()
	defer server.Close()
	db.data["delete_me"] = "bye"

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/v1/kv/delete_me", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.NotContains(t, db.data, "delete_me")
}

func TestDeleteKV_KeyNotFound(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/v1/kv/missing", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestKV_MethodNotAllowed(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPatch, server.URL+"/v1/kv/foo", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestGetKV_EmptyKeyPath(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/kv/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var res types.BaseResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.False(t, res.Success)
	assert.Equal(t, "missing key", res.Message)
}

func TestListKeys(t *testing.T) {
	server, db := setupTestServer()
	defer server.Close()

	db.data["a"] = "1"
	db.data["b"] = "2"

	resp, err := http.Get(server.URL + "/v1/keys")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp types.ListKeysResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))
	assert.True(t, listResp.Success)
	assert.ElementsMatch(t, []string{"a", "b"}, listResp.Keys)
}

func TestListKeys_DBError(t *testing.T) {
	server, db := setupTestServer()
	defer server.Close()
	db.errLst = errors.New("list fail")

	resp, err := http.Get(server.URL + "/v1/keys")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var res types.BaseResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.False(t, res.Success)
	assert.Equal(t, "Internal Server Error", res.Message)
}

func TestStatsEndpoint(t *testing.T) {
	server, db := setupTestServer()
	defer server.Close()
	db.stats = store.Stats{TotalKeys: 3, TotalSize: 100, Segments: 2}

	resp, err := http.Get(server.URL + "/v1/stats")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statsResp types.StatsResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&statsResp))
	assert.True(t, statsResp.Success)
	assert.Equal(t, 3, statsResp.TotalKeys)
	assert.Equal(t, int64(100), statsResp.TotalSize)
	assert.Equal(t, 2, statsResp.Segments)
}

func TestStatsEndpoint_DBError(t *testing.T) {
	server, db := setupTestServer()
	defer server.Close()
	db.errSta = errors.New("stats fail")

	resp, err := http.Get(server.URL + "/v1/stats")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var res types.BaseResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.False(t, res.Success)
	assert.Equal(t, "Internal Server Error", res.Message)
}

func TestNewHTTPServer_DefaultAddr(t *testing.T) {
	os.Unsetenv("LOGKV_ADDR")
	mux := http.NewServeMux()
	server := server.NewHTTPServer(mux)
	assert.Equal(t, ":8080", server.Addr)
}

func TestRegisterHooksLifecycle(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mux := http.NewServeMux()
	srv := &http.Server{Addr: "127.0.0.1:0", Handler: mux}

	mockLC := fxt.NewLifecycle(t)
	server.RegisterHooks(mockLC, srv, logger)

	ctx := context.Background()
	assert.NoError(t, mockLC.Start(ctx))
	assert.NoError(t, mockLC.Stop(ctx))
}
