package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricFileSaveRepository struct {
	mu         sync.RWMutex
	pathToFile string
}

func NewMetricFileSaveRepository(pathToFile string) *MetricFileSaveRepository {
	return &MetricFileSaveRepository{
		pathToFile: pathToFile,
	}
}

func (r *MetricFileSaveRepository) Save(ctx context.Context, metric types.Metrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(r.pathToFile), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(r.pathToFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(metric); err != nil {
		return err
	}

	return nil
}
