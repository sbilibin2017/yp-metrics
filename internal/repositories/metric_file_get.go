package repositories

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricFileGetRepository struct {
	mu         sync.RWMutex
	pathToFile string
}

func NewMetricFileGetRepository(pathToFile string) *MetricFileGetRepository {
	return &MetricFileGetRepository{pathToFile: pathToFile}
}

func (r *MetricFileGetRepository) Get(
	ctx context.Context,
	id types.MetricID,
) (*types.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, err := os.Open(r.pathToFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var metricFound *types.Metrics

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()

		var metric types.Metrics
		if err := json.Unmarshal(line, &metric); err != nil {
			continue
		}

		if metric.ID == id.ID && metric.MType == id.MType {
			metricFound = &metric
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return metricFound, nil
}
