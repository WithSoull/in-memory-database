package entry

import (
	"encoding/binary"
	"errors"
)

/*
8B   LSN
1B   CommandType
4B   KeyLen
4B   ValueLen
NB   Key
MB   Value
*/
const HeaderSize = 17 // 8B LSN + 1B CommandType + 4B KeyLen + 4B ValueLen

type Entry struct {
	LSN         uint64
	CommandType uint8
	Key         string
	Value       string
}

func (e *Entry) Encode() []byte {
	key := []byte(e.Key)
	value := []byte(e.Value)

	size := 8 + 1 + 4 + 4 + len(key) + len(value)

	buf := make([]byte, size)
	pos := 0

	binary.LittleEndian.PutUint64(buf[pos:], e.LSN)
	pos += 8

	buf[pos] = e.CommandType
	pos++

	binary.LittleEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	binary.LittleEndian.PutUint32(buf[pos:], uint32(len(value)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	copy(buf[pos:], value)

	return buf
}

func Decode(buf []byte) (Entry, error) {
	if len(buf) < HeaderSize {
		return Entry{}, errors.New("buffer too small")
	}

	pos := 0

	lsn := binary.LittleEndian.Uint64(buf[pos:])
	pos += 8

	cmd := buf[pos]
	pos++

	keyLen := binary.LittleEndian.Uint32(buf[pos:])
	pos += 4

	valLen := binary.LittleEndian.Uint32(buf[pos:])
	pos += 4

	total := pos + int(keyLen) + int(valLen)
	if total > len(buf) {
		return Entry{}, errors.New("buffer truncated")
	}

	key := string(buf[pos : pos+int(keyLen)])
	pos += int(keyLen)

	val := string(buf[pos : pos+int(valLen)])

	return Entry{
		LSN:         lsn,
		CommandType: cmd,
		Key:         key,
		Value:       val,
	}, nil
}
