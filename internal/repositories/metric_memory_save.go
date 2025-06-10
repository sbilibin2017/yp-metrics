package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricMemorySaveRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricMemorySaveRepository(
	data map[types.MetricID]types.Metrics,
) *MetricMemorySaveRepository {
	return &MetricMemorySaveRepository{data: data}
}

func (r *MetricMemorySaveRepository) Save(
	ctx context.Context,
	metrics []types.Metrics,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, metric := range metrics {
		r.data[types.MetricID{ID: metric.ID, MType: metric.MType}] = metric
	}

	return nil
}
