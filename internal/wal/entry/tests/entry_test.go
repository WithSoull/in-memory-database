package tests

import (
	"testing"

	"github.com/WithSoull/in-memory-database/internal/wal/entry"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   entry.Entry
	}{
		{
			name: "SET command",
			in: entry.Entry{
				LSN:         1,
				CommandType: 0x01,
				Key:         "user:123",
				Value:       "John",
			},
		},
		{
			name: "DEL command with empty value",
			in: entry.Entry{
				LSN:         42,
				CommandType: 0x02,
				Key:         "session:abc",
				Value:       "",
			},
		},
		{
			name: "empty key and value",
			in: entry.Entry{
				LSN:         0,
				CommandType: 0x01,
				Key:         "",
				Value:       "",
			},
		},
		{
			name: "large LSN",
			in: entry.Entry{
				LSN:         ^uint64(0),
				CommandType: 0x01,
				Key:         "k",
				Value:       "v",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := tt.in.Encode()
			got, err := entry.Decode(buf)

			require.NoError(t, err)
			require.Equal(t, tt.in, got)
		})
	}
}

func TestEncodeSize(t *testing.T) {
	t.Parallel()

	e := entry.Entry{
		LSN:         1,
		CommandType: 0x01,
		Key:         "abc",
		Value:       "defgh",
	}

	buf := e.Encode()
	expected := entry.HeaderSize + len("abc") + len("defgh")
	require.Equal(t, expected, len(buf))
}

func TestDecodeBufferTooSmall(t *testing.T) {
	t.Parallel()

	buf := make([]byte, entry.HeaderSize-1)
	_, err := entry.Decode(buf)

	require.Error(t, err)
	require.ErrorContains(t, err, "buffer too small")
}

func TestDecodeBufferTruncated(t *testing.T) {
	t.Parallel()

	e := entry.Entry{
		LSN:         1,
		CommandType: 0x01,
		Key:         "hello",
		Value:       "world",
	}

	buf := e.Encode()
	truncated := buf[:len(buf)-3]

	_, err := entry.Decode(truncated)

	require.Error(t, err)
	require.ErrorContains(t, err, "buffer truncated")
}

func TestDecodeExactHeaderOnly(t *testing.T) {
	t.Parallel()

	e := entry.Entry{
		LSN:         5,
		CommandType: 0x02,
		Key:         "",
		Value:       "",
	}

	buf := e.Encode()
	require.Equal(t, entry.HeaderSize, len(buf))

	got, err := entry.Decode(buf)

	require.NoError(t, err)
	require.Equal(t, e, got)
}

func TestDecodeIgnoresTrailingBytes(t *testing.T) {
	t.Parallel()

	e := entry.Entry{
		LSN:         1,
		CommandType: 0x01,
		Key:         "k",
		Value:       "v",
	}

	buf := e.Encode()
	buf = append(buf, 0xFF, 0xFF, 0xFF)

	got, err := entry.Decode(buf)

	require.NoError(t, err)
	require.Equal(t, e, got)
}
