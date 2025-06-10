package repositories

import (
	"context"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricMemoryGetRepository_GetByID(t *testing.T) {

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	data := map[types.MetricID]types.Metrics{
		{ID: "metric1", MType: types.Gauge}: {
			ID: "metric1", MType: types.Gauge, Value: float64Ptr(42.0),
		},
		{ID: "metric2", MType: types.Counter}: {
			ID: "metric2", MType: types.Counter, Delta: int64Ptr(10),
		},
	}

	repo := NewMetricMemoryGetRepository(data)

	t.Run("existing metric gauge", func(t *testing.T) {
		metric, err := repo.Get(context.Background(), types.MetricID{ID: "metric1", MType: types.Gauge})

		assert.NoError(t, err)
		assert.NotNil(t, metric)
		assert.Equal(t, "metric1", metric.ID)
		assert.Equal(t, types.Gauge, metric.MType)
		assert.NotNil(t, metric.Value)
		assert.Equal(t, 42.0, *metric.Value)
	})

	t.Run("existing metric counter", func(t *testing.T) {
		metric, err := repo.Get(context.Background(), types.MetricID{ID: "metric2", MType: types.Counter})

		assert.NoError(t, err)
		assert.NotNil(t, metric)
		assert.Equal(t, "metric2", metric.ID)
		assert.Equal(t, types.Counter, metric.MType)
		assert.NotNil(t, metric.Delta)
		assert.Equal(t, int64(10), *metric.Delta)
	})

	t.Run("non-existing metric", func(t *testing.T) {
		metric, err := repo.Get(context.Background(), types.MetricID{ID: "metric3", MType: types.Gauge})

		assert.NoError(t, err)
		assert.Nil(t, metric)
	})
}
