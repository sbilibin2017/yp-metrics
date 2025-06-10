package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricMemoryGetRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricMemoryGetRepository(
	data map[types.MetricID]types.Metrics,
) *MetricMemoryGetRepository {
	return &MetricMemoryGetRepository{data: data}
}

func (r *MetricMemoryGetRepository) Get(
	ctx context.Context,
	id types.MetricID,
) (*types.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metric, ok := r.data[id]
	if !ok {
		return nil, nil
	}

	return &metric, nil
}
