package services

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

// MetricSaver defines the interface for saving metrics.
type MetricSaver interface {
	// Save persists a slice of metrics.
	Save(ctx context.Context, metrics []types.Metrics) error
}

// MetricUpdateService provides metric update operations.
type MetricUpdateService struct {
	saver MetricSaver
}

// NewMetricUpdateService creates a new MetricUpdateService instance.
//
// Parameters:
//   - saver: an implementation of MetricSaver to persist metrics.
//
// Returns:
//
//	A pointer to a new MetricUpdateService.
func NewMetricUpdateService(
	saver MetricSaver,
) *MetricUpdateService {
	return &MetricUpdateService{saver: saver}
}

// Update processes a slice of metrics by accumulating counters and replacing gauges,
// then saves the normalized metrics via the underlying MetricSaver.
//
// Parameters:
//   - ctx: the context for controlling cancellation and deadlines.
//   - metrics: a slice of metrics to update.
//
// Returns:
//
//	An error if saving the metrics fails, otherwise nil.
func (svc *MetricUpdateService) Update(
	ctx context.Context,
	metrics []types.Metrics,
) error {
	accumulated := make(map[types.MetricID]types.Metrics)

	for _, m := range metrics {
		key := types.MetricID{ID: m.ID, MType: m.MType}

		switch m.MType {
		case types.Gauge:
			// For gauges, replace existing metric with latest value
			accumulated[key] = m
		case types.Counter:
			// For counters, accumulate values
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
