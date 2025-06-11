package runners

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
)

func RunWorker(ctx context.Context, worker func(ctx context.Context) error) {
	logger.Log.Info("Starting worker...")

	errChan := make(chan error, 1)

	go func() {
		errChan <- worker(ctx)
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("Context cancelled, stopping worker")
	case err := <-errChan:
		if err != nil {
			logger.Log.Errorf("Worker stopped with error: %v", err)
		} else {
			logger.Log.Info("Worker completed without error")
		}
	}
}
