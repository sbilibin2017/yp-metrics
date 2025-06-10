package runners

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
)

func RunWorker(ctx context.Context, worker func(ctx context.Context)) {
	logger.Log.Info("Starting worker...")

	go func() {
		worker(ctx)
	}()

	<-ctx.Done()

	logger.Log.Info("Context cancelled, stopping worker")

}
