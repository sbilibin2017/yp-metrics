package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type executor interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func getExecutor(
	ctx context.Context,
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) executor {
	if tx := txGetter(ctx); tx != nil {
		return tx
	}
	return db
}
