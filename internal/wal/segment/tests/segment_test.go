package segment_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WithSoull/in-memory-database/internal/wal/entry"
	"github.com/WithSoull/in-memory-database/internal/wal/segment"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newSegment(t *testing.T, maxSize int64) (*segment.Segment, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "wal_000001.log")
	s, err := segment.Open(path, maxSize, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s, path
}

func TestOpen_RequiresLogger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")
	_, err := segment.Open(path, 1024, nil)
	require.ErrorIs(t, err, segment.ErrInvalidLogger)
}

func TestWriteBatch_PersistsBytes(t *testing.T) {
	s, path := newSegment(t, 1<<20)

	entries := []entry.Entry{
		{LSN: 1, CommandType: entry.CommandSet, Key: "a", Value: "1"},
		{LSN: 2, CommandType: entry.CommandSet, Key: "b", Value: "2"},
	}
	require.NoError(t, s.WriteBatch(entries))

	stat, err := os.Stat(path)
	require.NoError(t, err)
	require.Greater(t, stat.Size(), int64(0))
	require.EqualValues(t, stat.Size(), s.Size())
}

func TestWriteBatch_Empty(t *testing.T) {
	s, _ := newSegment(t, 1<<20)
	require.NoError(t, s.WriteBatch(nil))
	require.EqualValues(t, 0, s.Size())
}

func TestReadAll_RoundTrip(t *testing.T) {
	s, _ := newSegment(t, 1<<20)
	want := []entry.Entry{
		{LSN: 1, CommandType: entry.CommandSet, Key: "a", Value: "alpha"},
		{LSN: 2, CommandType: entry.CommandDel, Key: "b", Value: ""},
		{LSN: 3, CommandType: entry.CommandSet, Key: "c", Value: "gamma"},
	}
	require.NoError(t, s.WriteBatch(want))

	got, err := s.ReadAll()
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestReadAll_StopsAtCorruption(t *testing.T) {
	s, path := newSegment(t, 1<<20)
	require.NoError(t, s.WriteBatch([]entry.Entry{
		{LSN: 1, CommandType: entry.CommandSet, Key: "a", Value: "1"},
		{LSN: 2, CommandType: entry.CommandSet, Key: "b", Value: "2"},
	}))
	require.NoError(t, s.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	data[len(data)-1] ^= 0xFF
	require.NoError(t, os.WriteFile(path, data, 0o644))

	reopened, err := segment.Open(path, 1<<20, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { _ = reopened.Close() })

	got, err := reopened.ReadAll()
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.EqualValues(t, 1, got[0].LSN)
}

func TestReadAll_AppendsAfterRead(t *testing.T) {
	s, _ := newSegment(t, 1<<20)
	require.NoError(t, s.WriteBatch([]entry.Entry{
		{LSN: 1, CommandType: entry.CommandSet, Key: "a", Value: "1"},
	}))

	_, err := s.ReadAll()
	require.NoError(t, err)

	require.NoError(t, s.WriteBatch([]entry.Entry{
		{LSN: 2, CommandType: entry.CommandSet, Key: "b", Value: "2"},
	}))

	got, err := s.ReadAll()
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.EqualValues(t, 2, got[1].LSN)
}

func TestIsFull(t *testing.T) {
	s, _ := newSegment(t, 64)
	require.False(t, s.IsFull())

	for i := range 5 {
		require.NoError(t, s.WriteBatch([]entry.Entry{
			{LSN: uint64(i), CommandType: entry.CommandSet, Key: "kkkk", Value: "vvvv"},
		}))
	}
	require.True(t, s.IsFull())
}
