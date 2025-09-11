package txIDctx_test

import (
	"context"
	"testing"

	txidctx "github.com/WithSoull/in-memory-database/internal/contextx/txIDctx"
	"github.com/stretchr/testify/require"
)

func TestInjectAndExtractTxID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ctx      context.Context
		txID     int64
		inject   bool
		expected int64
		ok       bool
	}{
		{
			name:     "successfully injected and extracted",
			ctx:      context.Background(),
			txID:     42,
			inject:   true,
			expected: 42,
			ok:       true,
		},
		{
			name:     "no TxID in context",
			ctx:      context.Background(),
			inject:   false,
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := tt.ctx
			if tt.inject {
				ctx = txidctx.InjectTxID(ctx, tt.txID)
			}

			id, ok := txidctx.ExtractIP(ctx)

			require.Equal(t, tt.expected, id)
			require.Equal(t, tt.ok, ok)
		})
	}
}
