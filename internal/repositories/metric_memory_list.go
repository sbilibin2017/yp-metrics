package repositories

import (
	"context"
	"sort"
	"sync"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricMemoryListRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

func NewMetricMemoryListRepository(
	data map[types.MetricID]types.Metrics,
) *MetricMemoryListRepository {
	return &MetricMemoryListRepository{data: data}
}

func (r *MetricMemoryListRepository) List(
	ctx context.Context,
) ([]types.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]types.Metrics, 0, len(r.data))
	for _, metric := range r.data {
		metrics = append(metrics, metric)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil
}
