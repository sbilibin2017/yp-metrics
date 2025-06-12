package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricDBSaveRepository struct {
	db       *sqlx.DB
	txGetter func(ctx context.Context) *sqlx.Tx
}

func NewMetricDBSaveRepository(
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) *MetricDBSaveRepository {
	return &MetricDBSaveRepository{db: db, txGetter: txGetter}
}

func (r *MetricDBSaveRepository) Save(
	ctx context.Context,
	metrics types.Metrics,
) error {
	exec := getExecutor(ctx, r.db, r.txGetter)

	_, err := exec.ExecContext(
		ctx,
		metricSaveQuery,
		metrics.ID,
		metrics.MType,
		metrics.Delta,
		metrics.Value,
	)

	return err
}

const metricSaveQuery = `
INSERT INTO content.metrics (id, mtype, delta, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, mtype) DO UPDATE SET
	delta = EXCLUDED.delta,
	value = EXCLUDED.value
`
