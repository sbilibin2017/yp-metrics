package workers

import (
	"context"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricsMemorySaver interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricsFileSaver interface {
	Save(ctx context.Context, metric types.Metrics) error
}

type MetricsMemoryLister interface {
	List(ctx context.Context) ([]types.Metrics, error)
}

type MetricsFileLister interface {
	List(ctx context.Context) ([]types.Metrics, error)
}

func StartMetricServerWorker(
	ctx context.Context,
	ms MetricsMemorySaver,
	fs MetricsFileSaver,
	ml MetricsMemoryLister,
	fl MetricsFileLister,
	storeInterval int,
	restore bool,
) {
	if restore {
		logger.Log.Info("Restoring metrics from file...")
		loadMetricsFromFile(ctx, fl, ms)
	}

	if storeInterval == 0 {
		logger.Log.Info("storeInterval = 0, saving metrics on shutdown only.")
		<-ctx.Done()
		saveMetricsToFile(ctx, ml, fs)
		return
	} else {
		logger.Log.Infof("Starting periodic saving every %d seconds", storeInterval)
		ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("Context canceled, saving metrics before shutdown...")
				saveMetricsToFile(ctx, ml, fs)
				return
			case <-ticker.C:
				logger.Log.Debug("Timer tick: saving metrics to file...")
				saveMetricsToFile(ctx, ml, fs)
			}
		}
	}
}

func loadMetricsFromFile(ctx context.Context, fl MetricsFileLister, ms MetricsMemorySaver) {
	metrics, err := fl.List(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to restore metrics from file: %v", err)
		return
	}

	logger.Log.Infof("Restoring %d metrics from file", len(metrics))

	for _, m := range metrics {
		if err := ms.Save(ctx, m); err != nil {
			logger.Log.Errorf("Failed to restore metric %s: %v", m.ID, err)
		} else {
			logger.Log.Debugf("Restored metric %s", m.ID)
		}
	}
}

func saveMetricsToFile(ctx context.Context, ml MetricsMemoryLister, fs MetricsFileSaver) {
	metrics, err := ml.List(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to list metrics: %v", err)
		return
	}

	logger.Log.Infof("Saving %d metrics to file", len(metrics))

	for _, m := range metrics {
		if err := fs.Save(ctx, m); err != nil {
			logger.Log.Errorf("Failed to save metric %s: %v", m.ID, err)
		} else {
			logger.Log.Debugf("Saved metric %s", m.ID)
		}
	}
}
