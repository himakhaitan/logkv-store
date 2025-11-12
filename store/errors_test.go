package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorInstantiation(t *testing.T) {
	t.Parallel()
	t.Run("All Errors Are Not Nil", func(t *testing.T) {
		assert.NotNil(t, ErrKeyNotFound, "ErrKeyNotFound must be initialized")
		assert.NotNil(t, ErrInvalidEntry, "ErrInvalidEntry must be initialized")
		assert.NotNil(t, ErrSegmentClosed, "ErrSegmentClosed must be initialized")
		assert.NotNil(t, ErrSegmentFull, "ErrSegmentFull must be initialized")
	})
}

func TestErrKeyNotFound_Message(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "key not found", ErrKeyNotFound.Error(), "ErrKeyNotFound message should match")
}

func TestErrInvalidEntry_Type(t *testing.T) {
	t.Parallel()
	var err error = ErrInvalidEntry
	assert.True(t, errors.Is(err, ErrInvalidEntry), "ErrInvalidEntry must be identifiable via errors.Is")
}

func TestSegmentErrors_Messages(t *testing.T) {
	t.Parallel()
	t.Run("Segment Closed Message", func(t *testing.T) {
		assert.Equal(t, "segment is closed", ErrSegmentClosed.Error(), "ErrSegmentClosed message should match")
	})

	t.Run("Segment Full Message", func(t *testing.T) {
		assert.Equal(t, "segment is full", ErrSegmentFull.Error(), "ErrSegmentFull message should match")
	})
}
