package services

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricSaver interface {
	Save(ctx context.Context, metrics []types.Metrics) error
}

type MetricUpdateService struct {
	saver MetricSaver
}

func NewMetricUpdateService(
	saver MetricSaver,
) *MetricUpdateService {
	return &MetricUpdateService{saver: saver}
}

func (svc *MetricUpdateService) Update(
	ctx context.Context,
	metrics []types.Metrics,
) error {
	accumulated := make(map[types.MetricID]types.Metrics)

	for _, m := range metrics {
		key := types.MetricID{ID: m.ID, MType: m.MType}

		switch m.MType {
		case types.Gauge:
			accumulated[key] = m
		case types.Counter:
			existing, ok := accumulated[key]
			if !ok {
				accumulated[key] = m
			} else {
				var sum int64
				if existing.Delta != nil {
					sum += *existing.Delta
				}
				if m.Delta != nil {
					sum += *m.Delta
				}

				accumulated[key] = types.Metrics{
					ID:    m.ID,
					MType: m.MType,
					Delta: &sum,
				}
			}
		}
	}

	normalized := make([]types.Metrics, 0, len(accumulated))
	for _, v := range accumulated {
		normalized = append(normalized, v)
	}

	return svc.saver.Save(ctx, normalized)
}
