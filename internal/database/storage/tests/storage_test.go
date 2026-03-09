package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"

	txidctx "github.com/WithSoull/in-memory-database/internal/contextx/txIDctx"
	"github.com/WithSoull/in-memory-database/internal/database/storage"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type setCall struct {
	key     string
	value   string
	txID    int64
	hasTxID bool
}

type getCall struct {
	key     string
	txID    int64
	hasTxID bool
}

type delCall struct {
	key     string
	txID    int64
	hasTxID bool
}

type engineSpy struct {
	setCalls []setCall
	getCalls []getCall
	delCalls []delCall

	getValue string
	getFound bool
}

type concurrentEngineSpy struct {
	mu          sync.Mutex
	txIDs       []int64
	missingTxID bool
}

func (e *engineSpy) Set(ctx context.Context, key, value string) {
	txID, ok := txidctx.ExtractIP(ctx)
	e.setCalls = append(e.setCalls, setCall{
		key:     key,
		value:   value,
		txID:    txID,
		hasTxID: ok,
	})
}

func (e *engineSpy) Get(ctx context.Context, key string) (string, bool) {
	txID, ok := txidctx.ExtractIP(ctx)
	e.getCalls = append(e.getCalls, getCall{
		key:     key,
		txID:    txID,
		hasTxID: ok,
	})
	return e.getValue, e.getFound
}

func (e *engineSpy) Del(ctx context.Context, key string) {
	txID, ok := txidctx.ExtractIP(ctx)
	e.delCalls = append(e.delCalls, delCall{
		key:     key,
		txID:    txID,
		hasTxID: ok,
	})
}

func (e *concurrentEngineSpy) Set(ctx context.Context, key, value string) {
	e.recordTxID(ctx)
}

func (e *concurrentEngineSpy) Get(ctx context.Context, key string) (string, bool) {
	e.recordTxID(ctx)
	return "", false
}

func (e *concurrentEngineSpy) Del(ctx context.Context, key string) {
	e.recordTxID(ctx)
}

func (e *concurrentEngineSpy) recordTxID(ctx context.Context) {
	txID, ok := txidctx.ExtractIP(ctx)

	e.mu.Lock()
	defer e.mu.Unlock()

	if !ok {
		e.missingTxID = true
		return
	}

	e.txIDs = append(e.txIDs, txID)
}

func (e *concurrentEngineSpy) snapshot() ([]int64, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]int64, len(e.txIDs))
	copy(result, e.txIDs)

	return result, e.missingTxID
}

func TestNewStrorage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		engine      storage.Engine
		logger      *zap.Logger
		expectedErr error
	}{
		{
			name:        "nil engine",
			engine:      nil,
			logger:      zap.NewNop(),
			expectedErr: derrors.ErrIvalidEngine,
		},
		{
			name:        "nil logger",
			engine:      &engineSpy{},
			logger:      nil,
			expectedErr: derrors.ErrInvalidLogger,
		},
		{
			name:   "success",
			engine: &engineSpy{},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := storage.NewStrorage(tt.engine, tt.logger)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				require.Nil(t, s)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, s)
		})
	}
}

func TestStorageOperationsInjectTxID(t *testing.T) {
	t.Parallel()

	engine := &engineSpy{
		getValue: "value",
		getFound: true,
	}

	s, err := storage.NewStrorage(engine, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, s.Set(ctx, "key", "value"))

	value, err := s.Get(ctx, "key")
	require.NoError(t, err)
	require.Equal(t, "value", value)

	require.NoError(t, s.Del(ctx, "key"))

	require.Len(t, engine.setCalls, 1)
	require.Equal(t, "key", engine.setCalls[0].key)
	require.Equal(t, "value", engine.setCalls[0].value)
	require.True(t, engine.setCalls[0].hasTxID)
	require.Equal(t, int64(1), engine.setCalls[0].txID)

	require.Len(t, engine.getCalls, 1)
	require.Equal(t, "key", engine.getCalls[0].key)
	require.True(t, engine.getCalls[0].hasTxID)
	require.Equal(t, int64(2), engine.getCalls[0].txID)

	require.Len(t, engine.delCalls, 1)
	require.Equal(t, "key", engine.delCalls[0].key)
	require.True(t, engine.delCalls[0].hasTxID)
	require.Equal(t, int64(3), engine.delCalls[0].txID)

	_, hasTxID := txidctx.ExtractIP(ctx)
	require.False(t, hasTxID)
}

func TestStorageGetNotFound(t *testing.T) {
	t.Parallel()

	engine := &engineSpy{
		getFound: false,
	}

	s, err := storage.NewStrorage(engine, zap.NewNop())
	require.NoError(t, err)

	value, err := s.Get(context.Background(), "missing")
	require.ErrorIs(t, err, derrors.ErrKeyNotFound)
	require.Empty(t, value)
	require.Len(t, engine.getCalls, 1)
	require.True(t, engine.getCalls[0].hasTxID)
	require.Equal(t, int64(1), engine.getCalls[0].txID)
}

func TestStorageConcurrentTxIDUniqueness(t *testing.T) {
	t.Parallel()

	const (
		workers       = 50
		opsPerWorker  = 200
		totalRequests = workers * opsPerWorker
	)

	engine := &concurrentEngineSpy{}
	s, err := storage.NewStrorage(engine, zap.NewNop())
	require.NoError(t, err)

	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	wg.Add(workers)

	for worker := 0; worker < workers; worker++ {
		worker := worker
		go func() {
			defer wg.Done()

			for i := 0; i < opsPerWorker; i++ {
				key := fmt.Sprintf("key-%d-%d", worker, i)
				value := fmt.Sprintf("value-%d-%d", worker, i)
				if setErr := s.Set(context.Background(), key, value); setErr != nil {
					errCh <- setErr
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for setErr := range errCh {
		require.NoError(t, setErr)
	}

	txIDs, missingTxID := engine.snapshot()
	require.False(t, missingTxID, "every call should have txID in context")
	require.Len(t, txIDs, totalRequests)

	seen := make(map[int64]struct{}, totalRequests)
	for _, txID := range txIDs {
		require.Greater(t, txID, int64(0))
		seen[txID] = struct{}{}
	}

	require.Len(t, seen, totalRequests, "txIDs should be unique")

	for expected := 1; expected <= totalRequests; expected++ {
		_, ok := seen[int64(expected)]
		require.True(t, ok, "missing txID=%d", expected)
	}
}
