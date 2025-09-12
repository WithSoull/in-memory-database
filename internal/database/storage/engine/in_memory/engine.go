package inmemory

import (
	"context"

	"github.com/WithSoull/in-memory-database/internal/database/storage"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"go.uber.org/zap"
)

type InMemoryEngine struct {
	hashtable Hashtable
	logger    *zap.Logger
}

func NewEngine(logger *zap.Logger) (storage.Engine, error) {
	return NewEngineWithHashtable(logger, NewHashtable())
}

func NewEngineWithHashtable(logger *zap.Logger, ht Hashtable) (storage.Engine, error) {
	if logger == nil {
		return nil, derrors.ErrInvalidLogger
	}
	if ht == nil {
		ht = NewHashtable()
	}
	return &InMemoryEngine{
		logger:    logger,
		hashtable: ht,
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
