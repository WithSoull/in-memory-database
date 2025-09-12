package storage

import (
	"context"

	txidctx "github.com/WithSoull/in-memory-database/internal/contextx/txIDctx"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"go.uber.org/zap"
)

type Storage interface {
	Set(context.Context, string, string) error
	Get(context.Context, string) (string, error)
	Del(context.Context, string) error
}

type storage struct {
	engine    Engine
	generator *IDGenerator
	logger    *zap.Logger
}

func NewStrorage(engine Engine, logger *zap.Logger) (Storage, error) {
	if engine == nil {
		return nil, derrors.ErrIvalidEngine
	}

	if logger == nil {
		return nil, derrors.ErrInvalidLogger
	}

	generator := NewIDGenerator()

	return &storage{
		engine:    engine,
		generator: generator,
		logger:    logger,
	}, nil
}

func (s *storage) Set(ctx context.Context, key, value string) error {
	txID := s.generator.Generate()
	ctx = txidctx.InjectTxID(ctx, txID)

	s.engine.Set(ctx, key, value)

	return nil
}

func (s *storage) Get(ctx context.Context, key string) (string, error) {
	txID := s.generator.Generate()
	ctx = txidctx.InjectTxID(ctx, txID)

	value, found := s.engine.Get(ctx, key)
	if !found {
		return "", derrors.ErrKeyNotFound
	}

	return value, nil
}

func (s *storage) Del(ctx context.Context, key string) error {
	txID := s.generator.Generate()
	ctx = txidctx.InjectTxID(ctx, txID)

	s.engine.Del(ctx, key)
	return nil
}
