package store

import (
	"encoding/binary"
	"time"
)

// Entry represents a single entry in the append-only log
type Entry struct {
	Timestamp uint32 // Unix timestamp
	KeySize   uint32 // Size of the key in bytes
	ValueSize uint32 // Size of the value in bytes
	Key       []byte // Key data
	Value     []byte // Value data
}

// TombstoneEntry represents a deleted entry (tombstone)
func (e *Entry) TombstoneEntry() *Entry {
	return &Entry{
		Timestamp: uint32(time.Now().Unix()),
		KeySize:   e.KeySize,
		ValueSize: 0, // Zero value size indicates tombstone
		Key:       e.Key,
		Value:     nil,
	}
}

// IsTombstone checks if this entry is a tombstone (deleted entry)
func (e *Entry) IsTombstone() bool {
	return e.ValueSize == 0
}

// Size returns the total size of the entry in bytes
func (e *Entry) Size() int {
	return 12 + int(e.KeySize) + int(e.ValueSize) // 12 bytes for timestamp + keysize + valuesize
}

// Serialize converts the entry to bytes for writing to disk
func (e *Entry) Serialize() []byte {
	buf := make([]byte, e.Size())
	offset := 0

	// Write timestamp (4 bytes)
	binary.LittleEndian.PutUint32(buf[offset:], e.Timestamp)
	offset += 4

	// Write key size (4 bytes)
	binary.LittleEndian.PutUint32(buf[offset:], e.KeySize)
	offset += 4

	// Write value size (4 bytes)
	binary.LittleEndian.PutUint32(buf[offset:], e.ValueSize)
	offset += 4

	// Write key data
	copy(buf[offset:], e.Key)
	offset += int(e.KeySize)

	// Write value data
	if e.ValueSize > 0 {
		copy(buf[offset:], e.Value)
	}

	return buf
}

// DeserializeEntry creates an entry from bytes read from disk
func DeserializeEntry(data []byte) (*Entry, error) {
	if len(data) < 12 {
		return nil, ErrInvalidEntry
	}

	entry := &Entry{}

	// Read timestamp
	entry.Timestamp = binary.LittleEndian.Uint32(data[0:4])

	// Read key size
	entry.KeySize = binary.LittleEndian.Uint32(data[4:8])

	// Read value size
	entry.ValueSize = binary.LittleEndian.Uint32(data[8:12])

	// Validate sizes
	if int(entry.KeySize)+int(entry.ValueSize) != len(data)-12 {
		return nil, ErrInvalidEntry
	}

	// Read key data
	entry.Key = make([]byte, entry.KeySize)
	copy(entry.Key, data[12:12+entry.KeySize])

	// Read value data
	if entry.ValueSize > 0 {
		entry.Value = make([]byte, entry.ValueSize)
		copy(entry.Value, data[12+entry.KeySize:12+entry.KeySize+entry.ValueSize])
	}

	return entry, nil
}
