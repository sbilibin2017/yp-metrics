package contexts

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGetTxFromContext(t *testing.T) {
	ctx := context.Background()

	// Создаём фейковую транзакцию (можно заменить на реальный *sqlx.Tx если нужно)
	tx := &sqlx.Tx{} // или: tx := &fakeTx{}

	ctxWithTx := SetTxToContext(ctx, tx)

	got := GetTxFromContext(ctxWithTx)
	assert.NotNil(t, got)
	assert.Equal(t, tx, got)
}

func TestGetTxFromContext_NoTx(t *testing.T) {
	ctx := context.Background()

	got := GetTxFromContext(ctx)
	assert.Nil(t, got)
}
