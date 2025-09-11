package txidctx

import (
	"context"

	"github.com/WithSoull/in-memory-database/internal/contextx"
)

const TxIDKey contextx.CtxKey = "TxID"

func InjectTxID(ctx context.Context, TxID int64) context.Context {
	return context.WithValue(ctx, TxIDKey, TxID)
}

func ExtractIP(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(TxIDKey).(int64)
	if !ok {
		return 0, false
	}
	return id, true
}
