package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/himakhaitan/logkv-store/engine"
	"github.com/himakhaitan/logkv-store/types"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewMux constructs the HTTP mux with all routes
func NewMux(db *engine.DB, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Health Check Route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// GET or DELETE /v1/kv/{key}
	mux.HandleFunc("/v1/kv/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract key from URL
		key := r.URL.Path[len("/v1/kv/"):]
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "missing key", Timestamp: time.Now().Unix()})
			return
		}
		switch r.Method {
		case http.MethodGet:
			value, err := db.Get(key)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: err.Error(), Timestamp: time.Now().Unix()})
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(types.GetResponse{Key: key, Value: value, BaseResponse: types.BaseResponse{Success: true, Timestamp: time.Now().Unix(), Message: "key fetched successfully"}})
		case http.MethodDelete:
			if err := db.Delete(key); err != nil {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: err.Error(), Timestamp: time.Now().Unix()})
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "method not allowed", Timestamp: time.Now().Unix()})
			return
		}
	})

	// PUT/POST /v1/kv
	mux.HandleFunc("/v1/kv", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "Method not allowed", Timestamp: time.Now().Unix()})
			return
		}
		var req types.SetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "invalid json", Timestamp: time.Now().Unix()})
			return
		}
		if req.Key == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "missing key", Timestamp: time.Now().Unix()})
			return
		}

		if err := db.Set(req.Key, req.Value); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: err.Error(), Timestamp: time.Now().Unix()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// GET /v1/keys
	mux.HandleFunc("/v1/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "Method not allowed", Timestamp: time.Now().Unix()})
			return
		}
		keys, err := db.List()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "Internal Server Error", Timestamp: time.Now().Unix()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.ListKeysResponse{
			Keys: keys,
			BaseResponse: types.BaseResponse{
				Success:   true,
				Timestamp: time.Now().Unix(),
				Message:   "keys fetched successfully",
			},
		})
	})

	// GET /v1/stats
	mux.HandleFunc("/v1/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "Method not allowed", Timestamp: time.Now().Unix()})
			return
		}
		stats, err := db.Stats()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(types.BaseResponse{Success: false, Message: "Internal Server Error", Timestamp: time.Now().Unix()})
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
	})

	return mux
}

// NewHTTPServer constructs the http.Server with configured addr
func NewHTTPServer(mux *http.ServeMux) *http.Server {
	addr := os.Getenv("LOGKV_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return &http.Server{Addr: addr, Handler: mux}
}

// RegisterHooks starts and stops the server using fx Lifecycle
func RegisterHooks(lc fx.Lifecycle, server *http.Server, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting Append-only log based Key-Value store", zap.String("addr", server.Addr))
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("Server failed to start", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping LogKV Store server")
			return server.Shutdown(ctx)
		},
	})
}
