package tests

import (
	"sync"
	"testing"

	"github.com/WithSoull/in-memory-database/internal/database/storage"
	"github.com/stretchr/testify/require"
)

func TestIDGeneratorGenerate(t *testing.T) {
	t.Parallel()

	g := storage.NewIDGenerator()
	require.NotNil(t, g)

	require.Equal(t, int64(1), g.Generate())
	require.Equal(t, int64(2), g.Generate())
	require.Equal(t, int64(3), g.Generate())
}

func TestIDGeneratorInstancesAreIndependent(t *testing.T) {
	t.Parallel()

	g1 := storage.NewIDGenerator()
	g2 := storage.NewIDGenerator()

	require.Equal(t, int64(1), g1.Generate())
	require.Equal(t, int64(1), g2.Generate())
}

func TestIDGeneratorConcurrentGenerate(t *testing.T) {
	t.Parallel()

	const total = 10000

	g := storage.NewIDGenerator()
	results := make(chan int64, total)

	var wg sync.WaitGroup
	wg.Add(total)

	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			results <- g.Generate()
		}()
	}

	wg.Wait()
	close(results)

	seen := make(map[int64]struct{}, total)
	for id := range results {
		require.Greater(t, id, int64(0))
		seen[id] = struct{}{}
	}

	require.Len(t, seen, total)
	for expected := 1; expected <= total; expected++ {
		_, ok := seen[int64(expected)]
		require.True(t, ok)
	}
}
