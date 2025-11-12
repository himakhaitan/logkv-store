package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/himakhaitan/logkv-store/pkg/config"
	"go.uber.org/zap/zaptest"
)

func setupStoreIntegration(t *testing.T) (*Store, string) {
	logger := zaptest.NewLogger(t)
	tempDir, err := os.MkdirTemp("", "store_int_test")
	require.NoError(t, err)

	cfg := &config.Config{DataDir: tempDir}

	realSM, err := NewSegmentManager(cfg.DataDir)
	require.NoError(t, err)

	realHT := NewHashTable()

	store := &Store{
		basePath:       cfg.DataDir,
		hashTable:      realHT,
		logger:         logger,
		segmentManager: realSM,
	}

	return store, tempDir
}

func TestStore_SetGetIntegration(t *testing.T) {
	t.Parallel()
	store, tempDir := setupStoreIntegration(t)
	defer os.RemoveAll(tempDir)
	defer store.Close()

	key := "user_id"
	value := "12345"

	err := store.Set(key, value)
	assert.NoError(t, err, "Set should succeed")

	result, err := store.Get(key)
	assert.NoError(t, err, "Get should succeed")
	assert.Equal(t, value, result, "Retrieved value must match set value")
}

func TestStore_DeleteIntegration(t *testing.T) {
	t.Parallel()
	store, tempDir := setupStoreIntegration(t)
	defer os.RemoveAll(tempDir)
	defer store.Close()

	key := "old_data"
	value := "to_be_deleted"

	store.Set(key, value)
	_, err := store.Get(key)
	require.NoError(t, err)

	err = store.Delete(key)
	assert.NoError(t, err, "Delete should succeed")

	_, err = store.Get(key)
	assert.ErrorIs(t, err, ErrKeyNotFound, "Get after Delete should return ErrKeyNotFound")

	keys, _ := store.List()
	found := false
	for _, k := range keys {
		if k == key {
			found = true
			break
		}
	}
	assert.False(t, found, "Deleted key should not appear in List()")
}

func TestStore_Set_Error(t *testing.T) {
	t.Parallel()
	store, tempDir := setupStoreIntegration(t)
	defer os.RemoveAll(tempDir)
	defer store.Close()

	store.segmentManager = nil

	err := store.Set("foo", "bar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "store not properly initialized")
}

func TestStore_Get_KeyNotFound(t *testing.T) {
	t.Parallel()
	store, tempDir := setupStoreIntegration(t)
	defer os.RemoveAll(tempDir)
	defer store.Close()

	_, err := store.Get("unknown_key")
	assert.ErrorIs(t, err, ErrKeyNotFound)
}
func TestStore_Delete_KeyNotFound(t *testing.T) {
	t.Parallel()
	store, tempDir := setupStoreIntegration(t)
	defer os.RemoveAll(tempDir)
	defer store.Close()

	err := store.Delete("nonexistent")
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestStore_New_MkdirAllFail(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{DataDir: "/root/invalid_dir"}
	s, err := New(logger, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestStore_New_SegmentManagerFail(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{DataDir: "/root/invalid_dir"}
	s, err := New(logger, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Nil(t, s.segmentManager)
}

func TestStore_loadFromSegments_SegmentManagerNil(t *testing.T) {
	store := &Store{
		hashTable: NewHashTable(),
		logger:    zaptest.NewLogger(t),
	}

	err := store.loadFromSegments()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "segment manager is not initialized")
}

func TestStore_LoadSegment_Tombstone(t *testing.T) {
	t.Parallel()
	store, tempDir := setupStoreIntegration(t)
	defer os.RemoveAll(tempDir)
	defer store.Close()

	err := store.Set("foo", "bar")
	assert.NoError(t, err)

	err = store.Delete("foo")
	assert.NoError(t, err)

	reloadedStore, err := New(store.logger, &config.Config{DataDir: tempDir})
	assert.NoError(t, err)
	defer reloadedStore.Close()

	_, err = reloadedStore.Get("foo")
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestStore_Stats_NoSegmentManager(t *testing.T) {
	store := &Store{
		hashTable:      NewHashTable(),
		segmentManager: nil,
	}

	stats, err := store.Stats()
	assert.NoError(t, err)
	assert.Equal(t, 0, stats.Segments)
}
func TestStore_LoadFromSegments_WithValidData(t *testing.T) {
	t.Parallel()
	store, tempDir := setupStoreIntegration(t)
	defer os.RemoveAll(tempDir)
	defer store.Close()

	err := store.Set("hello", "world")
	require.NoError(t, err)

	newStore, err := New(store.logger, &config.Config{DataDir: tempDir})
	require.NoError(t, err)
	defer newStore.Close()

	val, err := newStore.Get("hello")
	assert.NoError(t, err)
	assert.Equal(t, "world", val)
}
func TestStore_Close_NoSegmentManager(t *testing.T) {
	store := &Store{
		hashTable:      NewHashTable(),
		segmentManager: nil,
	}

	err := store.Close()
	assert.NoError(t, err, "Close with nil segmentManager should not fail")
}
