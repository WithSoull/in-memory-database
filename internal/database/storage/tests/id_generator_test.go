package tests

import (
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
