package contexts

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txKeyType struct{}

var txKey = txKeyType{}

func SetTxToContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func GetTxFromContext(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txKey).(*sqlx.Tx)
	return tx
}
