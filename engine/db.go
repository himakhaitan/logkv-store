package engine

import (
	"sync"

	"github.com/himakhaitan/logkv-store/store"
)

type DB struct {
	Store *store.Store
	mu    sync.RWMutex
}

func NewDB(s *store.Store) *DB {
	return &DB{Store: s}
}

func (db *DB) Get(key string) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.Store.Get(key)
}

func (db *DB) Set(key, value string) error {
	return db.Store.Set(key, value)
}

func (db *DB) Delete(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.Store.Delete(key)
}

func (db *DB) List() ([]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.Store.List()
}

func (db *DB) Stats() (store.Stats, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.Store.Stats()
}
