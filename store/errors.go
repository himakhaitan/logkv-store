package store

import "errors"

var (
	// ErrKeyNotFound is returned when a key is not found
	ErrKeyNotFound = errors.New("key not found")

	// ErrInvalidEntry is returned when an entry cannot be deserialized
	ErrInvalidEntry = errors.New("invalid entry")

	// ErrSegmentClosed is returned when trying to write to a closed segment
	ErrSegmentClosed = errors.New("segment is closed")

	// ErrSegmentFull is returned when a segment has reached its maximum size
	ErrSegmentFull = errors.New("segment is full")

	// ErrMergeInProgress prevents concurrent compactions.
	ErrMergeInProgress = errors.New("merge in progress")
)
