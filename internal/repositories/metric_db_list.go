package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricDBListRepository struct {
	db       *sqlx.DB
	txGetter func(ctx context.Context) *sqlx.Tx
}

func NewMetricDBListRepository(
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) *MetricDBListRepository {
	return &MetricDBListRepository{db: db, txGetter: txGetter}
}

func (r *MetricDBListRepository) List(ctx context.Context) ([]types.Metrics, error) {
	var metrics []types.Metrics

	exec := getExecutor(ctx, r.db, r.txGetter)

	err := exec.SelectContext(ctx, &metrics, metricListQuery)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

const metricListQuery = `
SELECT id, mtype, delta, value
FROM content.metrics
`
