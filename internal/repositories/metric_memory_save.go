package repositories

import (
	"context"
	"sync"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

// MetricMemorySaveRepository provides an in-memory storage for metrics.
//
// It uses a map keyed by MetricID to store metric values,
// and a RWMutex to synchronize concurrent access.
type MetricMemorySaveRepository struct {
	data map[types.MetricID]types.Metrics
	mu   sync.RWMutex
}

// NewMetricMemorySaveRepository creates a new MetricMemorySaveRepository instance.
//
// The provided data map is used as the underlying storage for metrics.
// It is recommended to pass an initialized map (e.g. make(map[types.MetricID]types.Metrics)).
//
// Returns a pointer to the newly created repository.
func NewMetricMemorySaveRepository(
	data map[types.MetricID]types.Metrics,
) *MetricMemorySaveRepository {
	return &MetricMemorySaveRepository{data: data}
}

// Save stores the given slice of metrics into the repository.
//
// It locks the repository for writing, then updates or inserts
// each metric into the internal data map, keyed by MetricID.
//
// Parameters:
//   - ctx: context for potential cancellation or deadlines (not currently used)
//   - metrics: slice of Metrics to be saved
//
// Returns an error if the operation fails (always returns nil currently).
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
