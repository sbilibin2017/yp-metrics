package repositories

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricFileListRepository struct {
	mu         sync.RWMutex
	pathToFile string
}

func NewMetricFileListRepository(pathToFile string) *MetricFileListRepository {
	return &MetricFileListRepository{pathToFile: pathToFile}
}

func (r *MetricFileListRepository) List(ctx context.Context) ([]types.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, err := os.Stat(r.pathToFile)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(r.pathToFile), 0755); err != nil {
			return nil, err
		}
		f, err := os.Create(r.pathToFile)
		if err != nil {
			return nil, err
		}
		f.Close()
		return []types.Metrics{}, nil
	} else if err != nil {
		return nil, err
	}

	file, err := os.Open(r.pathToFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	resultMap := make(map[types.MetricID]types.Metrics)

	for {
		var m types.Metrics
		if err := decoder.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		resultMap[types.MetricID{ID: m.ID, MType: m.MType}] = m
	}

	result := make([]types.Metrics, 0, len(resultMap))
	for _, m := range resultMap {
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}
