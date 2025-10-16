package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// SegmentManager manages multiple segments in the append-only log
type SegmentManager struct {
	mu       sync.RWMutex
	basePath string
	segments map[int]*Segment
	activeID int
	nextID   int
}

// NewSegmentManager creates a new segment manager
func NewSegmentManager(basePath string) (*SegmentManager, error) {
	sm := &SegmentManager{
		basePath: basePath,
		segments: make(map[int]*Segment),
		nextID:   1,
	}

	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Load existing segments
	if err := sm.loadSegments(); err != nil {
		return nil, fmt.Errorf("failed to load segments: %w", err)
	}

	// Create active segment if none exists
	if sm.activeID == 0 {
		if err := sm.createActiveSegment(); err != nil {
			return nil, fmt.Errorf("failed to create active segment: %w", err)
		}
	}

	return sm, nil
}

// loadSegments scans the base directory for existing segment files
func (sm *SegmentManager) loadSegments() error {
	files, err := filepath.Glob(filepath.Join(sm.basePath, "segment_*.log"))
	if err != nil {
		return fmt.Errorf("failed to scan for segment files: %w", err)
	}

	segmentIDs := make([]int, 0, len(files))
	segmentMap := make(map[int]*Segment)

	// Parse segment IDs and open segments
	for _, file := range files {
		var id int
		_, err := fmt.Sscanf(filepath.Base(file), "segment_%d.log", &id)
		if err != nil {
			continue // Skip invalid files
		}

		segment, err := OpenSegment(id, sm.basePath)
		if err != nil {
			return fmt.Errorf("failed to open segment %d: %w", id, err)
		}

		segmentIDs = append(segmentIDs, id)
		segmentMap[id] = segment

		// Track the highest ID
		if id >= sm.nextID {
			sm.nextID = id + 1
		}
	}

	// Sort segment IDs
	sort.Ints(segmentIDs)

	// Determine active segment (the last one that's not full)
	for i := len(segmentIDs) - 1; i >= 0; i-- {
		id := segmentIDs[i]
		segment := segmentMap[id]

		if segment.IsActive() {
			sm.activeID = id
			break
		}
	}

	sm.segments = segmentMap
	return nil
}

// createActiveSegment creates a new active segment
func (sm *SegmentManager) createActiveSegment() error {
	segment, err := NewSegment(sm.nextID, sm.basePath)
	if err != nil {
		return fmt.Errorf("failed to create new segment: %w", err)
	}

	sm.segments[sm.nextID] = segment
	sm.activeID = sm.nextID
	sm.nextID++

	return nil
}

// GetActiveSegment returns the currently active segment
func (sm *SegmentManager) GetActiveSegment() (*Segment, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.activeID == 0 {
		return nil, fmt.Errorf("no active segment")
	}

	segment, exists := sm.segments[sm.activeID]
	if !exists {
		return nil, fmt.Errorf("active segment %d not found", sm.activeID)
	}

	return segment, nil
}

// GetSegment returns a segment by ID
func (sm *SegmentManager) GetSegment(id int) (*Segment, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	segment, exists := sm.segments[id]
	return segment, exists
}

// Append writes an entry to the active segment
func (sm *SegmentManager) Append(entry *Entry) (int, int64, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Access active segment directly while holding write lock (avoid nested locks)
	if sm.activeID == 0 {
		return 0, 0, fmt.Errorf("no active segment")
	}
	segment, exists := sm.segments[sm.activeID]
	if !exists {
		return 0, 0, fmt.Errorf("active segment %d not found", sm.activeID)
	}

	offset, err := segment.Append(entry)
	if err != nil {
		if err == ErrSegmentFull {
			// Create new active segment
			if err := sm.createActiveSegment(); err != nil {
				return 0, 0, err
			}

			// Try again with new segment
			segment = sm.segments[sm.activeID]
			offset, err = segment.Append(entry)
			if err != nil {
				return 0, 0, err
			}
		} else {
			return 0, 0, err
		}
	}

	return segment.ID(), offset, nil
}

// Read reads an entry from a specific segment and position
func (sm *SegmentManager) Read(segmentID int, pos int64) (*Entry, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	segment, exists := sm.segments[segmentID]
	if !exists {
		return nil, fmt.Errorf("segment %d not found", segmentID)
	}

	return segment.Read(pos)
}

// Close closes all segments
func (sm *SegmentManager) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var lastErr error
	for _, segment := range sm.segments {
		if err := segment.Close(); err != nil {
			lastErr = err
		}
	}

	sm.segments = make(map[int]*Segment)
	sm.activeID = 0

	return lastErr
}

// GetSegmentIDs returns all segment IDs
func (sm *SegmentManager) GetSegmentIDs() []int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]int, 0, len(sm.segments))
	for id := range sm.segments {
		ids = append(ids, id)
	}

	sort.Ints(ids)
	return ids
}

// GetSegmentIDs returns all segment IDs
func (sm *SegmentManager) GetInactiveSegmentIDs() []int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]int, 0, len(sm.segments))
	for id, segment := range sm.segments {
		if !segment.isActive {
			ids = append(ids, id)
		}
	}

	sort.Ints(ids)
	return ids
}

// DeleteSegment delete a segment by ID
func (sm *SegmentManager) DeleteSegment(id int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	segment, exists := sm.segments[id]
	if !exists {
		return nil
	}

	delete(sm.segments, id)
	if err := segment.Delete(); err != nil {
		return err
	}
	return nil
}

// MergeFrom copies segment pointers from src into sm.
func (sm *SegmentManager) Merge(src *SegmentManager) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for k, v := range src.segments {
		sm.segments[k] = v
	}
}

// FlushAll fsyncs all segment files in the manager.
func (sm *SegmentManager) FlushAll() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, seg := range sm.segments {
		if err := seg.Flush(); err != nil {
			return err
		}
	}
	return nil
}
