package services

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
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
	metric, err := svc.getter.Get(ctx, id)
	if err != nil {
		logger.Log.Errorw("Failed to get metric", "id", id.ID, "type", id.MType, "error", err)
		return nil, types.ErrInternalServerError
	}
	if metric == nil {
		return nil, types.ErrMetricNotFound
	}
	return metric, nil
}
