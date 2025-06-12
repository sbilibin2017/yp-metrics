package repositories

import (
	"context"
	"errors"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type Lister interface {
	List(ctx context.Context) ([]types.Metrics, error)
}

type MetricListerContext struct {
	strategy Lister
}

func NewMetricListerContext() *MetricListerContext {
	return &MetricListerContext{}
}

func (m *MetricListerContext) SetContext(s Lister) {
	m.strategy = s
}

func (m *MetricListerContext) List(ctx context.Context) ([]types.Metrics, error) {
	if m.strategy == nil {
		return nil, errors.New("strategy is not set")
	}
	return m.strategy.List(ctx)
}
