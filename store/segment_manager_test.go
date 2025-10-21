package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type segmentTestContext struct {
	tempDir string
}

func setupTest(t *testing.T) *segmentTestContext {
	tempDir, err := os.MkdirTemp("", "sm_test")
	assert.NoError(t, err)

	return &segmentTestContext{
		tempDir: tempDir,
	}
}

func teardownTest(ctx *segmentTestContext) {
	os.RemoveAll(ctx.tempDir)
}

func createEntry(key string) *Entry {
	keyBytes := []byte(key)
	valueBytes := []byte(key + "_value")
	return &Entry{
		Timestamp: uint32(time.Now().Unix()),
		KeySize:   uint32(len(keyBytes)),
		ValueSize: uint32(len(valueBytes)),
		Key:       keyBytes,
		Value:     valueBytes,
	}
}

func TestNewSegmentManager_EmptyDir(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm, err := NewSegmentManager(ctx.tempDir)
	assert.NoError(t, err)

	assert.Equal(t, 1, sm.activeID, "Active ID should be 1")
	assert.Equal(t, 2, sm.nextID, "Next ID should be 2")
	assert.Len(t, sm.segments, 1, "Should have 1 segment")

	segment1 := sm.segments[1]
	assert.True(t, segment1.IsActive(), "The segment should be active")

	_, err = os.Stat(filepath.Join(ctx.tempDir, "segment_1.log"))
	assert.NoError(t, err, "Segment file should be created")
}

func TestNewSegmentManager_LoadExisting(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	file1Path := filepath.Join(ctx.tempDir, "segment_1.log")
	os.WriteFile(file1Path, []byte("some data"), 0644)

	file5Path := filepath.Join(ctx.tempDir, "segment_5.log")
	os.WriteFile(file5Path, []byte("data"), 0644)

	sm, err := NewSegmentManager(ctx.tempDir)
	assert.NoError(t, err)

	assert.Len(t, sm.segments, 3, "Should load 1, 5 and create 6")
	assert.Equal(t, 6, sm.activeID, "Active ID should be 6 (newly created)")
	assert.Equal(t, 7, sm.nextID, "Next ID should be 7")
	assert.Equal(t, []int{1, 5, 6}, sm.GetSegmentIDs(), "Segment IDs should be sorted")
}

func TestSegmentManager_AppendAndRead(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm, _ := NewSegmentManager(ctx.tempDir)
	entry1 := createEntry("key_1")
	entry2 := createEntry("key_2")

	segID1, offset1, err := sm.Append(entry1)
	assert.NoError(t, err)
	assert.Equal(t, 1, segID1)

	segID2, offset2, err := sm.Append(entry2)
	assert.NoError(t, err)
	assert.Equal(t, 1, segID2)
	assert.Greater(t, offset2, offset1, "Offset of second entry should be greater")

	readEntry1, err := sm.Read(segID1, offset1)
	assert.NoError(t, err, "Should be able to read the first entry")
	assert.NotNil(t, readEntry1)
	assert.Equal(t, string(entry1.Key), string(readEntry1.Key), "Read key must match")

	readEntry2, err := sm.Read(segID2, offset2)
	assert.NoError(t, err, "Should be able to read the second entry")
	assert.NotNil(t, readEntry2)
	assert.Equal(t, string(entry2.Key), string(readEntry2.Key), "Read key must match")

	_, err = sm.Read(segID1, 99999)
	assert.Error(t, err, "Should fail to read beyond segment size")

	_, err = sm.Read(99, 0)
	assert.ErrorContains(t, err, "segment 99 not found")
}

func TestSegmentManager_Append_SegmentSwitch_Forced(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm, _ := NewSegmentManager(ctx.tempDir)

	segment1 := sm.segments[1]

	segment1.mu.Lock()
	segment1.size = segment1.maxSize // Set size to MAX_SIZE to trip the full check
	segment1.entryCount = segment1.maxEntries
	segment1.mu.Unlock()

	entry := createEntry("small_trigger")

	// Append should now fail on Segment 1, trigger the switch, and succeed on Segment 2.
	segID, _, err := sm.Append(entry)

	assert.NoError(t, err)
	assert.Equal(t, 2, segID, "Should switch to segment 2")

	segment2, exists := sm.segments[2]
	assert.True(t, exists, "Segment 2 must exist after the switch")

	assert.False(t, segment1.IsActive(), "Segment 1 should now be inactive/full")
	assert.True(t, segment2.IsActive(), "Segment 2 should be the new active segment")
	assert.Equal(t, 2, sm.activeID, "Active ID should be 2 after switch")
	assert.Len(t, sm.segments, 2, "Should have two segments (1 and 2)")
	assert.GreaterOrEqual(t, segment2.EntryCount(), 1, "Segment 2 should have 1 entry")
}

func TestSegmentManager_Close(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm, _ := NewSegmentManager(ctx.tempDir)
	sm.Append(createEntry("data1"))
	sm.Append(createEntry("data2"))

	err := sm.Close()
	assert.NoError(t, err)

	assert.Equal(t, 0, sm.activeID, "Active ID should be 0 after Close")
	assert.Empty(t, sm.segments, "Segments map should be empty after Close")

	_, err = sm.Read(1, 0)
	assert.ErrorContains(t, err, "segment 1 not found", "Read should fail because segments map is empty")
}
func TestNewSegmentManager_OpenSegmentFails(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	os.RemoveAll(ctx.tempDir)
	err := os.WriteFile(ctx.tempDir, []byte("not a directory"), 0644)
	assert.NoError(t, err)

	_, err = NewSegmentManager(ctx.tempDir)
	assert.Error(t, err, "Expected error because base path is not a directory")
}

func TestSegmentManager_GetActiveSegment_NoActive(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm := &SegmentManager{
		basePath: ctx.tempDir,
		segments: make(map[int]*Segment),
		activeID: 0,
		nextID:   1,
	}

	_, err := sm.GetActiveSegment()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active segment")
}

func TestSegmentManager_Append_NoActive(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm := &SegmentManager{
		basePath: ctx.tempDir,
		segments: make(map[int]*Segment),
		activeID: 0,
		nextID:   1,
	}

	_, _, err := sm.Append(createEntry("k"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active segment")
}

func TestSegmentManager_Append_ActiveNotFound(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm := &SegmentManager{
		basePath: ctx.tempDir,
		segments: make(map[int]*Segment),
		activeID: 5,
		nextID:   6,
	}

	_, _, err := sm.Append(createEntry("k"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active segment 5 not found")
}

func TestSegmentManager_Read_MissingSegment(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	sm, _ := NewSegmentManager(ctx.tempDir)
	_, ok := sm.GetSegment(999)
	assert.False(t, ok)

	_, err := sm.Read(999, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "segment 999 not found")
}
