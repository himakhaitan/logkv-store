package store

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/himakhaitan/logkv-store/pkg/config"
	"go.uber.org/zap"
)

// Store represents a Bitcask-like append-only log key-value store
type Store struct {
	mu             sync.RWMutex
	basePath       string
	segmentManager *SegmentManager
	hashTable      *HashTable
	logger         *zap.Logger
}

// New creates a new Bitcask-like store
func New(logger *zap.Logger, config *config.Config) (*Store, error) {
	dataDir := config.DataDir
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Warn("Could not create data directory", zap.String("path", dataDir), zap.Error(err))
	}

	store := &Store{
		basePath:  dataDir,
		hashTable: NewHashTable(),
		logger:    logger,
	}

	// Initialize segment manager
	segmentManager, err := NewSegmentManager(dataDir)
	if err != nil {
		logger.Warn("Could not initialize segment manager", zap.String("path", dataDir), zap.Error(err))
		// Proceed without segment manager
		return store, nil
	}
	store.segmentManager = segmentManager

	// Load existing data from segments
	if err := store.loadFromSegments(); err != nil {
		logger.Error("Could not load data from segments", zap.String("path", dataDir), zap.Error(err))
		// End the store initialization if loading fails
		return nil, err
	}

	return store, nil
}

// loadFromSegments loads all existing data from segment files into the HashTable
func (s *Store) loadFromSegments() error {
	if s.segmentManager == nil {
		s.logger.Error("Segment manager is not initialized; cannot load segments")
		return fmt.Errorf("segment manager is not initialized")
	}

	segmentIDs := s.segmentManager.GetSegmentIDs()

	for _, segmentID := range segmentIDs {
		segment, exists := s.segmentManager.GetSegment(segmentID)
		if !exists {
			continue
		}

		// Read all entries from the segment
		if err := s.loadSegmentIntoKeyDir(segment); err != nil {
			s.logger.Error("Failed to load segment", zap.Int("segmentID", segmentID), zap.Error(err))
			return fmt.Errorf("failed to load segment %d: %w", segmentID, err)
		}
	}

	return nil
}

// loadSegmentIntoKeyDir loads all entries from a segment into the HashTable
func (s *Store) loadSegmentIntoKeyDir(segment *Segment) error {
	// For simplicity, we'll read from the beginning of the file
	// In a production system, you might want to maintain a more sophisticated index

	pos := int64(0)
	segmentSize := segment.Size()

	for pos < segmentSize {
		entry, err := segment.Read(pos)
		if err != nil {
			return fmt.Errorf("failed to read entry at position %d: %w", pos, err)
		}

		key := string(entry.Key)

		// Only add to HashTable if it's not a tombstone
		if !entry.IsTombstone() {
			s.hashTable.Put(key, segment.ID(), pos, entry.ValueSize, entry.Timestamp)
		} else {
			// Remove from HashTable if it's a tombstone
			s.hashTable.Delete(key)
		}

		// Move to next entry
		pos += int64(entry.Size())
	}

	return nil
}

// Get retrieves a value by key
func (s *Store) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.hashTable.Get(key)
	if !exists {
		return "", ErrKeyNotFound
	}

	// Read the entry from the segment
	logEntry, err := s.segmentManager.Read(entry.FileID, entry.ValuePos)
	if err != nil {
		return "", fmt.Errorf("failed to read entry: %w", err)
	}

	return string(logEntry.Value), nil
}

// Set stores a key-value pair
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Setting key:", key, "Value:", value)

	if s.segmentManager == nil {
		return fmt.Errorf("store not properly initialized")
	}

	// Create entry
	entry := &Entry{
		Timestamp: uint32(time.Now().Unix()),
		KeySize:   uint32(len(key)),
		ValueSize: uint32(len(value)),
		Key:       []byte(key),
		Value:     []byte(value),
	}

	// Append to active segment
	segmentID, offset, err := s.segmentManager.Append(entry)
	if err != nil {
		return fmt.Errorf("failed to append entry: %w", err)
	}

	// Update HashTable
	s.hashTable.Put(key, segmentID, offset, entry.ValueSize, entry.Timestamp)

	return nil
}

// Delete removes a key (creates a tombstone entry)
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.segmentManager == nil {
		return fmt.Errorf("store not properly initialized")
	}

	// Check if key exists
	_, exists := s.hashTable.Get(key)
	if !exists {
		return ErrKeyNotFound
	}

	// Create tombstone entry
	tombstoneEntry := &Entry{
		Timestamp: uint32(time.Now().Unix()),
		KeySize:   uint32(len(key)),
		ValueSize: 0, // Zero value size indicates tombstone
		Key:       []byte(key),
		Value:     nil,
	}

	// Append tombstone to active segment
	_, _, err := s.segmentManager.Append(tombstoneEntry)
	if err != nil {
		return fmt.Errorf("failed to append tombstone: %w", err)
	}

	// Remove from HashTable
	s.hashTable.Delete(key)

	return nil
}

// List returns all keys
func (s *Store) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.hashTable.List(), nil
}

// Stats returns database statistics
func (s *Store) Stats() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalKeys, totalSize := s.hashTable.Stats()

	// Count segments
	segmentCount := 0
	if s.segmentManager != nil {
		segmentCount = len(s.segmentManager.GetSegmentIDs())
	}

	return fmt.Sprintf("Total keys: %d\nTotal size: %d bytes\nSegments: %d",
		totalKeys, totalSize, segmentCount), nil
}

// Close closes the store and all its resources
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.segmentManager != nil {
		return s.segmentManager.Close()
	}

	return nil
}
