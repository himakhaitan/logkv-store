package store

import (
	"sync"
)

// HashTableEntry represents an entry in the HashTable for key lookups
type HashTableEntry struct {
	FileID    int    // ID of the segment file
	ValueSize uint32 // Size of the value
	ValuePos  int64  // Position of the value in the segment
	Timestamp uint32 // Timestamp when the entry was written
}

// HashTable is an in-memory hash index for key lookups
type HashTable struct {
	mu    sync.RWMutex
	index map[string]*HashTableEntry
}

// NewHashTable creates a new HashTable
func NewHashTable() *HashTable {
	return &HashTable{
		index: make(map[string]*HashTableEntry),
	}
}

// Put adds a key in the HashTable
func (kd *HashTable) Put(key string, fileID int, valuePos int64, valueSize uint32, timestamp uint32) {
	kd.mu.Lock()
	defer kd.mu.Unlock()

	kd.index[key] = &HashTableEntry{
		FileID:    fileID,
		ValueSize: valueSize,
		ValuePos:  valuePos,
		Timestamp: timestamp,
	}
}

// Get retrieves a key from the HashTable
func (kd *HashTable) Get(key string) (*HashTableEntry, bool) {
	kd.mu.RLock()
	defer kd.mu.RUnlock()

	entry, exists := kd.index[key]
	return entry, exists
}

// Delete removes a key from the HashTable
func (kd *HashTable) Delete(key string) {
	kd.mu.Lock()
	defer kd.mu.Unlock()

	delete(kd.index, key)
}

// List returns all keys in the HashTable
func (kd *HashTable) List() []string {
	kd.mu.RLock()
	defer kd.mu.RUnlock()

	keys := make([]string, 0, len(kd.index))
	for key := range kd.index {
		keys = append(keys, key)
	}
	return keys
}

// Stats returns statistics about the HashTable (optional)
func (kd *HashTable) Stats() (int, int64) {
	kd.mu.RLock()
	defer kd.mu.RUnlock()

	totalKeys := len(kd.index)
	totalSize := int64(0)

	for _, entry := range kd.index {
		totalSize += int64(entry.ValueSize)
	}

	return totalKeys, totalSize
}

// Merge applies updates from src only if current value still equals snap's.
// Prevents compaction from overwriting newer writes.
func (h *HashTable) Merge(src, snap *HashTable) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for k, v := range src.index {
		cur, ok := h.index[k]
		sv, okSnap := snap.index[k]
		// must exist in snapshot and be unchanged since snapshot
		if !okSnap || !ok || cur != sv {
			continue
		}
		h.index[k] = v
	}
}

// Clone returns a shallow snapshot of the table (for compaction checks).
func (h *HashTable) Clone() *HashTable {
	h.mu.RLock()
	defer h.mu.RUnlock()

	c := NewHashTable()
	for k, v := range h.index {
		c.index[k] = v
	}
	return c
}
