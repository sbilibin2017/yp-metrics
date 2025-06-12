package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricDBGetRepository struct {
	db       *sqlx.DB
	txGetter func(ctx context.Context) *sqlx.Tx
}

func NewMetricDBGetRepository(
	db *sqlx.DB,
	txGetter func(ctx context.Context) *sqlx.Tx,
) *MetricDBGetRepository {
	return &MetricDBGetRepository{db: db, txGetter: txGetter}
}

func (r *MetricDBGetRepository) Get(
	ctx context.Context,
	id types.MetricID,
) (*types.Metrics, error) {
	var metric types.Metrics

	exec := getExecutor(ctx, r.db, r.txGetter)

	err := exec.GetContext(ctx, &metric, metricGetQuery, id.ID, id.MType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &metric, nil
}

const metricGetQuery = `
SELECT id, mtype, delta, value
FROM content.metrics
WHERE id = $1 AND mtype = $2
`
