package inmemory

import (
	"context"

	"github.com/WithSoull/in-memory-database/internal/database/storage"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"go.uber.org/zap"
)

type InMemoryEngine struct {
	hashtable *Hashtable
	logger    *zap.Logger
}

func NewEngine(logger *zap.Logger) (storage.Engine, error) {
	if logger == nil {
		return nil, derrors.ErrInvalidLogger
	}
	return &InMemoryEngine{
		logger:    logger,
		hashtable: NewHashtable(),
	}, nil
}

func (e *InMemoryEngine) Set(ctx context.Context, key, value string) {
	e.logger.Debug("Set method")
	e.hashtable.Set(key, value)
}

func (e *InMemoryEngine) Get(ctx context.Context, key string) (string, bool) {
	e.logger.Debug("Get method")
	value, found := e.hashtable.Get(key)
	return value, found
}

func (e *InMemoryEngine) Del(ctx context.Context, key string) {
	e.logger.Debug("Del method")
	e.hashtable.Del(key)
}
