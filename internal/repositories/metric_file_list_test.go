package repositories

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFileListRepository_List(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "metrics_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Подготавливаем тестовые метрики
	v1 := float64(42)
	v2 := float64(100)
	v3 := int64(10)
	data := []types.Metrics{
		{ID: "metric1", MType: "gauge", Value: &v1},
		{ID: "metric2", MType: "counter", Delta: &v3},
		{ID: "metric1", MType: "gauge", Value: &v2}, // обновлённая метрика с тем же ID и MType
	}

	for _, m := range data {
		b, err := json.Marshal(m)
		require.NoError(t, err)
		_, err = tmpFile.Write(append(b, '\n'))
		require.NoError(t, err)
	}
	require.NoError(t, tmpFile.Close())

	repo := NewMetricFileListRepository(tmpFile.Name())

	result, err := repo.List(context.Background())
	require.NoError(t, err)

	// Ожидается 2 уникальные метрики
	assert.Len(t, result, 2)

	// Проверяем, что последняя версия metric1 перезаписала первую
	var foundGauge, foundCounter bool
	for _, m := range result {
		switch m.ID {
		case "metric1":
			foundGauge = true
			assert.Equal(t, "gauge", m.MType)
			assert.Equal(t, v2, *m.Value)
		case "metric2":
			foundCounter = true
			assert.Equal(t, "counter", m.MType)
			assert.Equal(t, v3, *m.Delta)
		}
	}
	assert.True(t, foundGauge)
	assert.True(t, foundCounter)
}
