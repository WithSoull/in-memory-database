package entry

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const (
	CommandSet uint8 = 1
	CommandDel uint8 = 2

	crcSize    = 4
	headerSize = crcSize + 8 + 1 + 4 + 4 // CRC + LSN + Command + KeyLen + ValueLen
)

var (
	ErrBufferTooSmall = errors.New("entry: buffer too small")
	ErrTruncated      = errors.New("entry: buffer truncated")
	ErrCRCMismatch    = errors.New("entry: crc mismatch")
)

type Entry struct {
	LSN         uint64
	CommandType uint8
	Key         string
	Value       string
}

func (e *Entry) Encode() []byte {
	key := []byte(e.Key)
	value := []byte(e.Value)

	buf := make([]byte, headerSize+len(key)+len(value))

	binary.LittleEndian.PutUint64(buf[crcSize:], e.LSN)
	buf[crcSize+8] = e.CommandType
	binary.LittleEndian.PutUint32(buf[crcSize+9:], uint32(len(key)))
	binary.LittleEndian.PutUint32(buf[crcSize+13:], uint32(len(value)))
	copy(buf[headerSize:], key)
	copy(buf[headerSize+len(key):], value)

	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE(buf[crcSize:]))

	return buf
}

func Decode(buf []byte) (Entry, error) {
	if len(buf) < headerSize {
		return Entry{}, ErrBufferTooSmall
	}

	keyLen := binary.LittleEndian.Uint32(buf[crcSize+9:])
	valueLen := binary.LittleEndian.Uint32(buf[crcSize+13:])
	total := headerSize + int(keyLen) + int(valueLen)
	if total > len(buf) {
		return Entry{}, ErrTruncated
	}

	wantCRC := binary.LittleEndian.Uint32(buf)
	gotCRC := crc32.ChecksumIEEE(buf[crcSize:total])
	if wantCRC != gotCRC {
		return Entry{}, ErrCRCMismatch
	}

	return Entry{
		LSN:         binary.LittleEndian.Uint64(buf[crcSize:]),
		CommandType: buf[crcSize+8],
		Key:         string(buf[headerSize : headerSize+keyLen]),
		Value:       string(buf[headerSize+keyLen : total]),
	}, nil
}

func EncodedSize(key, value string) int {
	return headerSize + len(key) + len(value)
}
