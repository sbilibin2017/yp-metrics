package repositories

import (
	"context"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricMemorySaveRepository_Save(t *testing.T) {
	ptrFloat64 := func(v float64) *float64 {
		return &v
	}

	ptrInt64 := func(v int64) *int64 {
		return &v
	}

	tests := []struct {
		name     string
		metrics  []types.Metrics
		expected map[types.MetricID]types.Metrics
	}{
		{
			name:     "empty metrics slice",
			metrics:  []types.Metrics{},
			expected: map[types.MetricID]types.Metrics{},
		},
		{
			name: "single gauge metric",
			metrics: []types.Metrics{
				{
					ID:    "g1",
					MType: "gauge",
					Value: ptrFloat64(99.9),
				},
			},
			expected: map[types.MetricID]types.Metrics{
				{ID: "g1", MType: "gauge"}: {
					ID:    "g1",
					MType: "gauge",
					Value: ptrFloat64(99.9),
				},
			},
		},
		{
			name: "gauge and counter",
			metrics: []types.Metrics{
				{
					ID:    "metric1",
					MType: "gauge",
					Value: ptrFloat64(42.0),
				},
				{
					ID:    "metric2",
					MType: "counter",
					Delta: ptrInt64(10),
				},
			},
			expected: map[types.MetricID]types.Metrics{
				{ID: "metric1", MType: "gauge"}: {
					ID:    "metric1",
					MType: "gauge",
					Value: ptrFloat64(42.0),
				},
				{ID: "metric2", MType: "counter"}: {
					ID:    "metric2",
					MType: "counter",
					Delta: ptrInt64(10),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialData := make(map[types.MetricID]types.Metrics)
			repo := NewMetricMemorySaveRepository(initialData)

			for _, m := range tt.metrics {
				err := repo.Save(context.Background(), m)
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, repo.data)
		})
	}
}
