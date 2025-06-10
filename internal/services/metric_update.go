package services

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricUpdateSaver interface {
	Save(ctx context.Context, metrics types.Metrics) error
}

type MetricUpdateGetter interface {
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

type MetricUpdateService struct {
	saver  MetricUpdateSaver
	getter MetricUpdateGetter
}

func NewMetricUpdateService(
	saver MetricUpdateSaver,
	getter MetricUpdateGetter,
) *MetricUpdateService {
	return &MetricUpdateService{saver: saver, getter: getter}
}

func (svc *MetricUpdateService) Update(
	ctx context.Context,
	metrics types.Metrics,
) error {
	if metrics.MType == types.Counter {
		currentMetric, err := svc.getter.Get(ctx, types.MetricID{ID: metrics.ID, MType: metrics.MType})
		if err != nil {
			return types.ErrInternalServerError
		}
		if currentMetric != nil {
			*metrics.Delta += *currentMetric.Delta
		}
	}
	return svc.saver.Save(ctx, metrics)
}
