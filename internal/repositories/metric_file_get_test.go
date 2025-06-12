package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricFileGetRepository_Get(t *testing.T) {
	float64Ptr := func(f float64) *float64 {
		return &f
	}

	metric := types.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: float64Ptr(123.45),
	}

	// Create temp file with the JSON of one metric
	tmpFile, err := os.CreateTemp("", "metric_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = json.NewEncoder(tmpFile).Encode(metric)
	assert.NoError(t, err)
	tmpFile.Close()

	repo := NewMetricFileGetRepository(tmpFile.Name())

	// Act
	got, err := repo.Get(context.Background(), types.MetricID{ID: "Alloc", MType: "gauge"})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, metric.ID, got.ID)
	assert.Equal(t, metric.MType, got.MType)
	assert.NotNil(t, got.Value)
	assert.Equal(t, *metric.Value, *got.Value)
}
