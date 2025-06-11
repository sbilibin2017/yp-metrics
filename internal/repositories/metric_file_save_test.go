package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFileSaveRepository_Save(t *testing.T) {
	// Создаём временный файл
	tmpFile, err := os.CreateTemp("", "metrics_test_*.jsonl")
	require.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	repo := NewMetricFileSaveRepository(tmpFilePath)
	ctx := context.Background()

	v := float64(42)
	metric := types.Metrics{
		ID:    "metric1",
		MType: "gauge",
		Value: &v,
	}

	// Сохраняем метрику
	err = repo.Save(ctx, metric)
	assert.NoError(t, err)

	// Читаем и проверяем содержимое файла
	data, err := os.ReadFile(tmpFilePath)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"id":"metric1"`)
	assert.Contains(t, string(data), `"value":42`)
}
