package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/himakhaitan/logkv-store/engine"
	"github.com/himakhaitan/logkv-store/types"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewMux constructs the HTTP mux with all routes
func NewMux(db *engine.DB, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// GET or DELETE /v1/kv/{key}
	mux.HandleFunc("/v1/kv/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[len("/v1/kv/"):]
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			value, err := db.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(types.GetResponse{Key: key, Value: value})
		case http.MethodDelete:
			if err := db.Delete(key); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// PUT/POST /v1/kv
	mux.HandleFunc("/v1/kv", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Incoming request:", r.Method)
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req types.SetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.Key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}
		log.Printf("Setting key %q to %q\n", req.Key, req.Value)

		if err := db.Set(req.Key, req.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Set complete")
		w.WriteHeader(http.StatusNoContent)
	})

	// GET /v1/keys
	mux.HandleFunc("/v1/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		keys, err := db.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(keys)
	})

	// GET /v1/stats
	mux.HandleFunc("/v1/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		stats, err := db.Stats()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(stats))
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
