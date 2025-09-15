package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	// DefaultMaxSegmentSize is the default maximum size for a segment file (10MB)
	DefaultMaxSegmentSize = 10 * 1024 * 1024

	// DefaultMaxEntriesPerSegment is the default maximum number of entries per segment
	DefaultMaxEntriesPerSegment = 10000
)

// Segment represents a single segment file in the append-only log
type Segment struct {
	mu         sync.RWMutex
	id         int
	path       string
	file       *os.File
	size       int64
	entryCount int
	maxSize    int64
	maxEntries int
	isActive   bool
	isClosed   bool
}

// NewSegment creates a new segment
func NewSegment(id int, basePath string) (*Segment, error) {
	path := filepath.Join(basePath, fmt.Sprintf("segment_%d.log", id))

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create segment file: %w", err)
	}

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat segment file: %w", err)
	}

	segment := &Segment{
		id:         id,
		path:       path,
		file:       file,
		size:       stat.Size(),
		maxSize:    DefaultMaxSegmentSize,
		maxEntries: DefaultMaxEntriesPerSegment,
		isActive:   true,
		isClosed:   false,
	}

	return segment, nil
}

// OpenSegment opens an existing segment for reading
func OpenSegment(id int, basePath string) (*Segment, error) {
	path := filepath.Join(basePath, fmt.Sprintf("segment_%d.log", id))

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open segment file: %w", err)
	}

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat segment file: %w", err)
	}

	segment := &Segment{
		id:         id,
		path:       path,
		file:       file,
		size:       stat.Size(),
		maxSize:    DefaultMaxSegmentSize,
		maxEntries: DefaultMaxEntriesPerSegment,
		isActive:   false,
		isClosed:   false,
	}

	return segment, nil
}

// Append writes an entry to the segment
func (s *Segment) Append(entry *Entry) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isClosed {
		return 0, ErrSegmentClosed
	}

	if !s.isActive {
		return 0, ErrSegmentClosed
	}

	// Check if segment is full
	if s.size >= s.maxSize || s.entryCount >= s.maxEntries {
		s.isActive = false
		return 0, ErrSegmentFull
	}

	// Serialize entry
	data := entry.Serialize()

	// Write to file
	offset := s.size
	_, err := s.file.Write(data)
	if err != nil {
		return 0, fmt.Errorf("failed to write entry: %w", err)
	}

	// Update segment stats
	s.size += int64(len(data))
	s.entryCount++

	return offset, nil
}

// Read reads an entry from the segment at the given position
func (s *Segment) Read(pos int64) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if pos >= s.size {
		return nil, fmt.Errorf("position %d is beyond segment size %d", pos, s.size)
	}

	// Seek to position
	_, err := s.file.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to position %d: %w", pos, err)
	}

	// Read entry header (12 bytes: timestamp + keysize + valuesize)
	header := make([]byte, 12)
	_, err = io.ReadFull(s.file, header)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry header: %w", err)
	}

	// Parse sizes
	keySize := binary.LittleEndian.Uint32(header[4:8])
	valueSize := binary.LittleEndian.Uint32(header[8:12])

	// Read full entry
	entrySize := 12 + int(keySize) + int(valueSize)
	entryData := make([]byte, entrySize)
	copy(entryData, header)

	_, err = io.ReadFull(s.file, entryData[12:])
	if err != nil {
		return nil, fmt.Errorf("failed to read entry data: %w", err)
	}

	return DeserializeEntry(entryData)
}

// Close closes the segment
func (s *Segment) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isClosed {
		return nil
	}

	s.isActive = false
	s.isClosed = true

	return s.file.Close()
}

// IsActive returns whether the segment is active (can be written to)
func (s *Segment) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isActive && !s.isClosed
}

// Size returns the current size of the segment
func (s *Segment) Size() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.size
}

// EntryCount returns the number of entries in the segment
func (s *Segment) EntryCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entryCount
}

// ID returns the segment ID
func (s *Segment) ID() int {
	return s.id
}

// Path returns the segment file path
func (s *Segment) Path() string {
	return s.path
}
