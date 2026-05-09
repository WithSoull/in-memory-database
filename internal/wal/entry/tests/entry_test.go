package entry_test

import (
	"encoding/binary"
	"testing"

	"github.com/WithSoull/in-memory-database/internal/wal/entry"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode_RoundTrip(t *testing.T) {
	cases := []entry.Entry{
		{LSN: 1, CommandType: entry.CommandSet, Key: "k", Value: "v"},
		{LSN: 42, CommandType: entry.CommandDel, Key: "key", Value: ""},
		{LSN: 1 << 40, CommandType: entry.CommandSet, Key: "", Value: ""},
		{LSN: 7, CommandType: entry.CommandSet, Key: "a/b/c", Value: "пример"},
	}

	for _, want := range cases {
		t.Run(want.Key, func(t *testing.T) {
			buf := want.Encode()
			got, err := entry.Decode(buf)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}

func TestDecode_BufferTooSmall(t *testing.T) {
	_, err := entry.Decode([]byte{0x01, 0x02})
	require.ErrorIs(t, err, entry.ErrBufferTooSmall)
}

func TestDecode_Truncated(t *testing.T) {
	e := entry.Entry{LSN: 1, CommandType: entry.CommandSet, Key: "key", Value: "value"}
	buf := e.Encode()

	_, err := entry.Decode(buf[:len(buf)-2])
	require.ErrorIs(t, err, entry.ErrTruncated)
}

func TestDecode_CRCMismatch(t *testing.T) {
	e := entry.Entry{LSN: 1, CommandType: entry.CommandSet, Key: "key", Value: "value"}
	buf := e.Encode()

	buf[len(buf)-1] ^= 0xFF

	_, err := entry.Decode(buf)
	require.ErrorIs(t, err, entry.ErrCRCMismatch)
}

func TestDecode_CRCMismatch_HeaderCorruption(t *testing.T) {
	e := entry.Entry{LSN: 1, CommandType: entry.CommandSet, Key: "key", Value: "value"}
	buf := e.Encode()

	binary.LittleEndian.PutUint64(buf[4:], 999)

	_, err := entry.Decode(buf)
	require.ErrorIs(t, err, entry.ErrCRCMismatch)
}
