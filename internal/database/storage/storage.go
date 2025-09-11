package storage

import (
	"context"

	txidctx "github.com/WithSoull/in-memory-database/internal/contextx/txIDctx"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"go.uber.org/zap"
)

type Storage struct {
	engine    Engine
	generator *IDGenerator
	logger    *zap.Logger
}

func NewStrorage(engine Engine, logger *zap.Logger) (*Storage, error) {
	if engine == nil {
		return nil, derrors.ErrIvalidEngine
	}

	if logger == nil {
		return nil, derrors.ErrInvalidLogger
	}

	generator := NewIDGenerator()

	return &Storage{
		engine:    engine,
		generator: generator,
		logger:    logger,
	}, nil
}

func (s *Storage) Set(ctx context.Context, key, value string) error {
	txID := s.generator.Generate()
	ctx = txidctx.InjectTxID(ctx, txID)

	s.engine.Set(ctx, key, value)

	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	txID := s.generator.Generate()
	ctx = txidctx.InjectTxID(ctx, txID)

	value, found := s.engine.Get(ctx, key)
	if !found {
		return "", derrors.ErrKeyNotFound
	}

	return value, nil
}

func (s *Storage) Del(ctx context.Context, key string) error {
	txID := s.generator.Generate()
	ctx = txidctx.InjectTxID(ctx, txID)

	s.engine.Del(ctx, key)
	return nil
}
