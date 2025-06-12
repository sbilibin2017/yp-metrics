package repositories

import (
	"context"
	"errors"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type Saver interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricSaverContext struct {
	strategy Saver
}

func NewMetricSaverContext() *MetricSaverContext {
	return &MetricSaverContext{}
}

func (m *MetricSaverContext) SetContext(s Saver) {
	m.strategy = s
}

func (m *MetricSaverContext) Save(ctx context.Context, metric types.Metrics) error {
	if m.strategy == nil {
		return errors.New("strategy is not set")
	}
	return m.strategy.Save(ctx, metric)
}
