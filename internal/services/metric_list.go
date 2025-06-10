package services

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricLister interface {
	List(ctx context.Context) ([]types.Metrics, error)
}

type MetricListService struct {
	lister MetricLister
}

func NewMetricListService(
	lister MetricLister,
) *MetricListService {
	return &MetricListService{lister: lister}
}

func (svc *MetricListService) List(
	ctx context.Context,
) ([]types.Metrics, error) {
	return svc.lister.List(ctx)
}
