package repositories

import (
	"context"
	"errors"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type Getter interface {
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

type MetricGetterContext struct {
	strategy Getter
}

func NewMetricGetterContext() *MetricGetterContext {
	return &MetricGetterContext{}
}

func (m *MetricGetterContext) SetContext(s Getter) {
	m.strategy = s
}

func (m *MetricGetterContext) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	if m.strategy == nil {
		return nil, errors.New("strategy is not set")
	}
	return m.strategy.Get(ctx, id)
}
