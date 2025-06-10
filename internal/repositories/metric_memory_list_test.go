package repositories_test

import (
	"context"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricMemoryListRepository_List(t *testing.T) {
	ctx := context.Background()

	counterValue := int64(10)
	gaugeValue := 3.14

	data := map[types.MetricID]types.Metrics{
		{ID: "metricB", MType: types.Counter}: {
			ID:    "metricB",
			MType: types.Counter,
			Delta: &counterValue,
		},
		{ID: "metricA", MType: types.Gauge}: {
			ID:    "metricA",
			MType: types.Gauge,
			Value: &gaugeValue,
		},
	}

	repo := repositories.NewMetricMemoryListRepository(data)

	metrics, err := repo.List(ctx)

	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

	assert.Equal(t, "metricA", metrics[0].ID)
	assert.Equal(t, "metricB", metrics[1].ID)

	assert.Equal(t, types.Gauge, metrics[0].MType)
	assert.NotNil(t, metrics[0].Value)
	assert.Equal(t, gaugeValue, *metrics[0].Value)

	assert.Equal(t, types.Counter, metrics[1].MType)
	assert.NotNil(t, metrics[1].Delta)
	assert.Equal(t, counterValue, *metrics[1].Delta)
}
