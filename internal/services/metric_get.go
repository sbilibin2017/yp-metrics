package services

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricGetter interface {
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

type MetricGetService struct {
	getter MetricGetter
}

func NewMetricGetService(
	getter MetricGetter,
) *MetricGetService {
	return &MetricGetService{getter: getter}
}

func (svc *MetricGetService) Get(
	ctx context.Context,
	id types.MetricID,
) (*types.Metrics, error) {
	return svc.getter.Get(ctx, id)
}
