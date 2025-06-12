package repositories

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/stretchr/testify/require"
)

func TestGetExecutor(t *testing.T) {
	ctx := context.Background()

	// Case 1: txGetter returns a non-nil *sqlx.Tx (fake with empty struct pointer)
	tx := &sqlx.Tx{}
	exec := getExecutor(ctx, nil, func(ctx context.Context) *sqlx.Tx {
		return tx
	})
	require.Equal(t, tx, exec)

	// Case 2: txGetter returns nil, so should return the db
	db := &sqlx.DB{}
	exec = getExecutor(ctx, db, func(ctx context.Context) *sqlx.Tx {
		return nil
	})
	require.Equal(t, db, exec)
}
