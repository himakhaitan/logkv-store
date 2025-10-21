package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/himakhaitan/logkv-store/pkg/config"
	"github.com/himakhaitan/logkv-store/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDBOperations(t *testing.T) {
	tempDir := t.TempDir()
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{DataDir: filepath.Join(tempDir, "data")}

	s, err := store.New(logger, cfg)
	assert.NoError(t, err)
	db := NewDB(s)

	// Test Set
	err = db.Set("foo", "bar")
	assert.NoError(t, err)

	// Test Get
	val, err := db.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", val)

	// Test List
	keys, err := db.List()
	assert.NoError(t, err)
	assert.Contains(t, keys, "foo")

	// Test Delete
	err = db.Delete("foo")
	assert.NoError(t, err)

	// After delete, should return key not found
	_, err = db.Get("foo")
	assert.Error(t, err)

	// Test Stats
	stats, err := db.Stats()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, stats.TotalKeys, 0)

	err = s.Close()
	assert.NoError(t, err)

	os.RemoveAll(tempDir)
}
