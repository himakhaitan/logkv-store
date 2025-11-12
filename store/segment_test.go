package store

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestEntry(key string, value string) *Entry {
	keyBytes := []byte(key)
	valueBytes := []byte(value)
	return &Entry{
		Timestamp: uint32(time.Now().Unix()),
		KeySize:   uint32(len(keyBytes)),
		ValueSize: uint32(len(valueBytes)),
		Key:       keyBytes,
		Value:     valueBytes,
	}
}

func TestNewSegment(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	segmentID := 1
	seg, err := NewSegment(segmentID, ctx.tempDir)
	assert.NoError(t, err)
	defer seg.Close()
	assert.Equal(t, segmentID, seg.ID())
	assert.True(t, seg.IsActive(), "New segment should be active")
	assert.Equal(t, int64(0), seg.Size())
}

func TestOpenSegment(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	segmentID := 2

	initSeg, _ := NewSegment(segmentID, ctx.tempDir)
	initSeg.Append(createTestEntry("k", "v"))
	initSize := initSeg.Size()
	initSeg.Close()

	seg, err := OpenSegment(segmentID, ctx.tempDir)
	assert.NoError(t, err)
	defer seg.Close()
	assert.False(t, seg.IsActive(), "Opened segment should be inactive/read-only")
	assert.Equal(t, initSize, seg.Size(), "Size should match file size")
}

func TestSegment_AppendAndRead(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	seg, _ := NewSegment(3, ctx.tempDir)
	defer seg.Close()

	entry1 := createTestEntry("key_1", "value_a")
	entry2 := createTestEntry("key_2", "value_b")
	entry1Size := int64(len(entry1.Serialize()))

	offset1, err1 := seg.Append(entry1)
	assert.NoError(t, err1)

	offset2, err2 := seg.Append(entry2)
	assert.NoError(t, err2)
	assert.Equal(t, entry1Size, offset2)

	readEntry1, err3 := seg.Read(offset1)
	assert.NoError(t, err3)
	assert.True(t, bytes.Equal(entry1.Value, readEntry1.Value))

	readEntry2, err4 := seg.Read(offset2)
	assert.NoError(t, err4)
	assert.True(t, bytes.Equal(entry2.Value, readEntry2.Value))
}

func TestSegment_FullCapacityChecks(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	entry := createTestEntry("a", "b")
	segSize, _ := NewSegment(4, ctx.tempDir)
	defer segSize.Close()

	segSize.mu.Lock()
	segSize.size = segSize.maxSize
	segSize.mu.Unlock()

	_, err := segSize.Append(entry)
	assert.ErrorIs(t, err, ErrSegmentFull)

	segCount, _ := NewSegment(5, ctx.tempDir)
	defer segCount.Close()

	segCount.mu.Lock()
	segCount.entryCount = segCount.maxEntries
	segCount.mu.Unlock()

	_, err = segCount.Append(entry)
	assert.ErrorIs(t, err, ErrSegmentFull)
}

func TestSegment_Close(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	seg, _ := NewSegment(6, ctx.tempDir)

	assert.NoError(t, seg.Close())

	_, err := seg.Append(createTestEntry("k", "v"))
	assert.ErrorIs(t, err, ErrSegmentClosed)
}

func TestSegment_ReadErrorConditions(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	seg, _ := NewSegment(7, ctx.tempDir)
	defer seg.Close()

	entry := createTestEntry("x", "y")
	seg.Append(entry)

	_, err := seg.Read(seg.Size() + 1)
	assert.ErrorContains(t, err, "is beyond segment size", "Read should fail beyond EOF")

	_, err = seg.Read(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read", "Read from invalid position should fail to read header or data")
}

func TestSegment_Concurrency(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	seg, _ := NewSegment(8, ctx.tempDir)
	defer seg.Close()

	var wg sync.WaitGroup
	numRoutines := 10
	numEntries := 5

	// Append and Read concurrently
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numEntries; j++ {
				key := fmt.Sprintf("k_%d_%d", id, j)
				entry := createTestEntry(key, "data")

				offset, err := seg.Append(entry)
				assert.NoError(t, err)

				readEntry, err := seg.Read(offset)
				assert.NoError(t, err)
				assert.True(t, bytes.Equal(entry.Key, readEntry.Key))
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, numRoutines*numEntries, seg.EntryCount(), "All entries must be counted")
	// The primary check here is that the test completes without data corruption or race errors.
}

func TestSegment_AccessorMethods(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	seg, _ := NewSegment(9, ctx.tempDir)
	defer seg.Close()

	seg.Append(createTestEntry("k", "v"))

	assert.Equal(t, 9, seg.ID())
	assert.True(t, seg.IsActive())
	assert.Greater(t, seg.Size(), int64(0))
	assert.Equal(t, 1, seg.EntryCount())
	assert.Equal(t, int64(DefaultMaxSegmentSize), seg.maxSize, "maxSize must be correctly set and match DefaultMaxSegmentSize")
	assert.Equal(t, DefaultMaxEntriesPerSegment, seg.maxEntries)
	assert.Contains(t, seg.Path(), "segment_9.log")
}
func TestSegment_ErrorPaths(t *testing.T) {
	t.Parallel()
	ctx := setupTest(t)
	defer teardownTest(ctx)

	t.Run("NewSegment fails on invalid path", func(t *testing.T) {
		_, err := NewSegment(1, "/invalid/path/that/does/not/exist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create segment file")
	})

	t.Run("NewSegment fails when file.Stat fails", func(t *testing.T) {
		dir := filepath.Join(ctx.tempDir, "unreadable_dir")
		err := os.Mkdir(dir, 0000)
		assert.NoError(t, err)
		_, err = NewSegment(2, dir)
		assert.Error(t, err, "Expected Stat to fail due to unreadable directory")
		os.Chmod(dir, 0755)
	})

	t.Run("OpenSegment fails for non-existent file", func(t *testing.T) {
		_, err := OpenSegment(9999, ctx.tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open segment file")
	})

	t.Run("Append fails when write fails", func(t *testing.T) {
		seg, _ := NewSegment(10, ctx.tempDir)
		seg.file.Close()
		entry := createTestEntry("bad", "write")

		_, err := seg.Append(entry)
		assert.ErrorContains(t, err, "failed to write entry")
	})

	t.Run("Read fails when seek fails", func(t *testing.T) {
		seg, _ := NewSegment(11, ctx.tempDir)
		defer seg.Close()

		entry := createTestEntry("a", "b")
		seg.Append(entry)
		seg.file.Close()

		_, err := seg.Read(0)
		assert.ErrorContains(t, err, "failed to seek")
	})

	t.Run("Read fails due to incomplete header", func(t *testing.T) {
		path := filepath.Join(ctx.tempDir, "segment_incomplete.log")
		f, _ := os.Create(path)
		f.Write([]byte{1, 2, 3, 4})
		f.Close()

		seg := &Segment{
			id:   42,
			path: path,
			file: mustOpenFile(path),
			size: 4,
		}

		_, err := seg.Read(0)
		assert.ErrorContains(t, err, "failed to read entry header")
		seg.Close()
	})

	t.Run("Read fails due to incomplete entry data", func(t *testing.T) {
		path := filepath.Join(ctx.tempDir, "segment_partial.log")
		f, _ := os.Create(path)
		header := make([]byte, 12)
		binary.LittleEndian.PutUint32(header[4:8], 4)
		binary.LittleEndian.PutUint32(header[8:12], 4)
		f.Write(header)
		f.Write([]byte("abcd"))
		f.Close()

		seg := &Segment{
			id:   99,
			path: path,
			file: mustOpenFile(path),
			size: 16,
		}

		_, err := seg.Read(0)
		assert.ErrorContains(t, err, "failed to read entry data")
		seg.Close()
	})

	t.Run("Close called twice", func(t *testing.T) {
		seg, _ := NewSegment(12, ctx.tempDir)
		assert.NoError(t, seg.Close())
		assert.NoError(t, seg.Close(), "Second close should be a no-op")
	})
}

// Helper to reopen a file safely
func mustOpenFile(path string) *os.File {
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	return f
}
