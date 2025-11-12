package store

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	key1       = "key_1"
	key2       = "key_2"
	fileID1    = 1
	fileID2    = 2
	valuePos1  = int64(100)
	valuePos2  = int64(200)
	valueSize1 = uint32(50)
	valueSize2 = uint32(75)
	timestamp1 = uint32(1678886400)
	timestamp2 = uint32(1678886500)
)

func TestNewHashTable(t *testing.T) {
	t.Parallel()
	ht := NewHashTable()
	assert.NotNil(t, ht, "NewHashTable should not return nil")
	assert.NotNil(t, ht.index, "Internal index map should be initialized")
	assert.Zero(t, len(ht.index), "New hash table should be empty")
}

func TestHashTable_PutAndGet(t *testing.T) {
	t.Parallel()
	ht := NewHashTable()

	ht.Put(key1, fileID1, valuePos1, valueSize1, timestamp1)
	entry, exists := ht.Get(key1)

	assert.True(t, exists, "Key should exist after Put")
	assert.Equal(t, fileID1, entry.FileID)
	assert.Equal(t, valuePos1, entry.ValuePos)
	assert.Equal(t, valueSize1, entry.ValueSize)
	assert.Equal(t, timestamp1, entry.Timestamp)

	ht.Put(key1, fileID2, valuePos2, valueSize2, timestamp2)
	updatedEntry, exists := ht.Get(key1)

	assert.True(t, exists, "Key should still exist after update")
	assert.Equal(t, fileID2, updatedEntry.FileID, "FileID should be updated")
	assert.Equal(t, valueSize2, updatedEntry.ValueSize, "ValueSize should be updated")
	assert.Equal(t, timestamp2, updatedEntry.Timestamp, "Timestamp should be updated")

	_, exists = ht.Get("missing_key")
	assert.False(t, exists, "Non-existent key should not be found")
}

func TestHashTable_Delete(t *testing.T) {
	t.Parallel()
	ht := NewHashTable()
	ht.Put(key1, fileID1, valuePos1, valueSize1, timestamp1)

	ht.Delete(key1)
	_, exists := ht.Get(key1)
	assert.False(t, exists, "Key should be deleted")

	assert.NotPanics(t, func() { ht.Delete(key2) }, "Deleting non-existent key should not panic")
}

func TestHashTable_List(t *testing.T) {
	t.Parallel()
	ht := NewHashTable()

	ht.Put(key1, fileID1, valuePos1, valueSize1, timestamp1)
	ht.Put(key2, fileID1, valuePos1, valueSize1, timestamp1)

	keys := ht.List()

	expectedKeys := []string{key1, key2}

	sort.Strings(keys)
	sort.Strings(expectedKeys)

	assert.ElementsMatch(t, expectedKeys, keys, "List should return all keys present in the map")

	ht.Delete(key1)
	ht.Delete(key2)
	assert.Empty(t, ht.List(), "List should return empty slice after all keys are deleted")
}

func TestHashTable_Stats(t *testing.T) {
	t.Parallel()
	ht := NewHashTable()

	count, size := ht.Stats()
	assert.Zero(t, count, "Initial key count should be 0")
	assert.Zero(t, size, "Initial total size should be 0")

	ht.Put(key1, fileID1, valuePos1, valueSize1, timestamp1) // size 50
	ht.Put(key2, fileID1, valuePos1, valueSize2, timestamp1) // size 75

	count, size = ht.Stats()
	assert.Equal(t, 2, count, "Key count should be 2")
	assert.Equal(t, int64(valueSize1+valueSize2), size, "Total size should be 50 + 75 = 125")

	ht.Put(key1, fileID2, valuePos2, uint32(100), timestamp2) // New size 100

	count, size = ht.Stats()
	assert.Equal(t, 2, count, "Key count should remain 2 after update")
	// Total size = 100 (key1 new size) + 75 (key2 size) = 175
	assert.Equal(t, int64(175), size, "Total size should reflect the updated entry size")

	ht.Delete(key2) // Remove size 75 entry

	count, size = ht.Stats()
	assert.Equal(t, 1, count, "Key count should be 1 after delete")
	// Total size = 100 (key1 size)
	assert.Equal(t, int64(100), size, "Total size should reflect the deletion")
}

func TestHashTable_Concurrency(t *testing.T) {
	ht := NewHashTable()
	numGoroutines := 100
	numOperationsPerGoroutine := 100

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Simultaneous reads and writes without data race or deadlock.

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				key := "key_" + time.Now().String()
				// Write (Lock) operation
				ht.Put(key, id, int64(j), uint32(j), uint32(time.Now().Unix()))
				// Delete (Lock) operation
				if j%10 == 0 {
					ht.Delete(key)
				}
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				// Read (RLock) operations
				ht.List()
				ht.Stats()
				ht.Get(key1)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.GreaterOrEqual(t, len(ht.index), 0, "Final index size should be valid")
	case <-time.After(5 * time.Second):
		t.Fatal("Concurrency test timed out (possible deadlock)")
	}
}
