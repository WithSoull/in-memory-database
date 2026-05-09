package segment

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/WithSoull/in-memory-database/internal/wal/entry"
	"go.uber.org/zap"
)

var ErrInvalidLogger = errors.New("segment: logger is required")

type Segment struct {
	file        *os.File
	currentSize int64
	maxSize     int64
	logger      *zap.Logger
}

func Open(path string, maxSize int64, logger *zap.Logger) (*Segment, error) {
	if logger == nil {
		return nil, ErrInvalidLogger
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("segment: open %q: %w", path, err)
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("segment: stat %q: %w", path, err)
	}

	return &Segment{
		file:        file,
		currentSize: stat.Size(),
		maxSize:     maxSize,
		logger:      logger,
	}, nil
}

func (s *Segment) WriteBatch(entries []entry.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	total := 0
	for i := range entries {
		total += entry.EncodedSize(entries[i].Key, entries[i].Value)
	}

	buf := make([]byte, 0, total)
	for i := range entries {
		buf = append(buf, entries[i].Encode()...)
	}

	n, err := s.file.Write(buf)
	if err != nil {
		return fmt.Errorf("segment: write: %w", err)
	}
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("segment: fsync: %w", err)
	}

	s.currentSize += int64(n)
	return nil
}

func (s *Segment) ReadAll() ([]entry.Entry, error) {
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("segment: seek: %w", err)
	}

	data, err := io.ReadAll(s.file)
	if err != nil {
		return nil, fmt.Errorf("segment: read: %w", err)
	}

	var entries []entry.Entry
	offset := 0
	for offset < len(data) {
		e, err := entry.Decode(data[offset:])
		if err != nil {
			s.logger.Warn("segment: stopping read at corrupted entry, returning valid prefix",
				zap.String("file", s.file.Name()),
				zap.Int("offset", offset),
				zap.Error(err),
			)
			break
		}
		entries = append(entries, e)
		offset += entry.EncodedSize(e.Key, e.Value)
	}

	if _, err := s.file.Seek(0, io.SeekEnd); err != nil {
		return nil, fmt.Errorf("segment: seek end: %w", err)
	}

	return entries, nil
}

func (s *Segment) IsFull() bool {
	return s.currentSize >= s.maxSize
}

func (s *Segment) Size() int64 {
	return s.currentSize
}

func (s *Segment) Close() error {
	return s.file.Close()
}
