package workers

import (
	"context"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type metricsUpdateResult struct {
	Request types.Metrics
	Err     error
}

type MetricsUpdater interface {
	Update(ctx context.Context, req types.Metrics) error
}

func StartMetricAgentWorker(
	ctx context.Context,
	metricsUpdater MetricsUpdater,
	pollInterval int,
	reportInterval int,
) {
	pollMetricsCh := pollMetrics(
		ctx,
		pollInterval,
		[]func() []types.Metrics{
			collectRuntimeCounterMetrics,
			collectRuntimeGaugeMetrics,
		},
	)
	reportMetricsCh := reportMetrics(
		ctx,
		metricsUpdater,
		reportInterval,
		pollMetricsCh,
	)
	logResults(ctx, reportMetricsCh)
}

func pollMetrics(ctx context.Context, pollInterval int, collectors []func() []types.Metrics) <-chan types.Metrics {
	metricsCh := make(chan types.Metrics)

	go func() {
		defer close(metricsCh)
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("pollMetricsLoop stopped due to context cancellation")
				return
			case <-ticker.C:
				for _, collector := range collectors {
					metrics := collector()
					logger.Log.Infof("Collected %d metrics", len(metrics))
					for _, m := range metrics {
						metricsCh <- m
					}
				}
			}
		}
	}()

	return metricsCh
}

func collectRuntimeGaugeMetrics() []types.Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	float64Ptr := func(f float64) *float64 {
		return &f
	}

	metrics := []types.Metrics{
		{ID: "Alloc", MType: types.Gauge, Value: float64Ptr(float64(memStats.Alloc))},
		{ID: "BuckHashSys", MType: types.Gauge, Value: float64Ptr(float64(memStats.BuckHashSys))},
		{ID: "Frees", MType: types.Gauge, Value: float64Ptr(float64(memStats.Frees))},
		{ID: "GCCPUFraction", MType: types.Gauge, Value: float64Ptr(memStats.GCCPUFraction)},
		{ID: "GCSys", MType: types.Gauge, Value: float64Ptr(float64(memStats.GCSys))},
		{ID: "HeapAlloc", MType: types.Gauge, Value: float64Ptr(float64(memStats.HeapAlloc))},
		{ID: "HeapIdle", MType: types.Gauge, Value: float64Ptr(float64(memStats.HeapIdle))},
		{ID: "HeapInuse", MType: types.Gauge, Value: float64Ptr(float64(memStats.HeapInuse))},
		{ID: "HeapObjects", MType: types.Gauge, Value: float64Ptr(float64(memStats.HeapObjects))},
		{ID: "HeapReleased", MType: types.Gauge, Value: float64Ptr(float64(memStats.HeapReleased))},
		{ID: "HeapSys", MType: types.Gauge, Value: float64Ptr(float64(memStats.HeapSys))},
		{ID: "LastGC", MType: types.Gauge, Value: float64Ptr(float64(memStats.LastGC))},
		{ID: "Lookups", MType: types.Gauge, Value: float64Ptr(float64(memStats.Lookups))},
		{ID: "MCacheInuse", MType: types.Gauge, Value: float64Ptr(float64(memStats.MCacheInuse))},
		{ID: "MCacheSys", MType: types.Gauge, Value: float64Ptr(float64(memStats.MCacheSys))},
		{ID: "MSpanInuse", MType: types.Gauge, Value: float64Ptr(float64(memStats.MSpanInuse))},
		{ID: "MSpanSys", MType: types.Gauge, Value: float64Ptr(float64(memStats.MSpanSys))},
		{ID: "Mallocs", MType: types.Gauge, Value: float64Ptr(float64(memStats.Mallocs))},
		{ID: "NextGC", MType: types.Gauge, Value: float64Ptr(float64(memStats.NextGC))},
		{ID: "NumForcedGC", MType: types.Gauge, Value: float64Ptr(float64(memStats.NumForcedGC))},
		{ID: "NumGC", MType: types.Gauge, Value: float64Ptr(float64(memStats.NumGC))},
		{ID: "OtherSys", MType: types.Gauge, Value: float64Ptr(float64(memStats.OtherSys))},
		{ID: "PauseTotalNs", MType: types.Gauge, Value: float64Ptr(float64(memStats.PauseTotalNs))},
		{ID: "StackInuse", MType: types.Gauge, Value: float64Ptr(float64(memStats.StackInuse))},
		{ID: "StackSys", MType: types.Gauge, Value: float64Ptr(float64(memStats.StackSys))},
		{ID: "Sys", MType: types.Gauge, Value: float64Ptr(float64(memStats.Sys))},
		{ID: "TotalAlloc", MType: types.Gauge, Value: float64Ptr(float64(memStats.TotalAlloc))},
		{ID: "RandomValue", MType: types.Gauge, Value: float64Ptr(rand.Float64() * 100)},
	}

	return metrics
}

func collectRuntimeCounterMetrics() []types.Metrics {
	val := int64(1)
	metrics := []types.Metrics{
		{ID: "PollCount", MType: types.Counter, Delta: &val},
	}
	return metrics
}

func reportMetrics(
	ctx context.Context,
	metricsUpdater MetricsUpdater,
	reportInterval int,
	in <-chan types.Metrics,
) <-chan metricsUpdateResult {
	out := make(chan metricsUpdateResult, 100)
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	go func() {
		defer func() {
			ticker.Stop()
			close(out)
		}()

		var buffer []types.Metrics

		for {
			select {
			case <-ctx.Done():
				for _, m := range buffer {
					err := metricsUpdater.Update(ctx, m)
					out <- metricsUpdateResult{Request: m, Err: err}
				}
				return

			case m, ok := <-in:
				if !ok {
					for _, mm := range buffer {
						err := metricsUpdater.Update(ctx, mm)
						out <- metricsUpdateResult{Request: mm, Err: err}
					}
					return
				}
				buffer = append(buffer, m)

			case <-ticker.C:
				for _, m := range buffer {
					err := metricsUpdater.Update(ctx, m)
					out <- metricsUpdateResult{Request: m, Err: err}
				}
				buffer = buffer[:0]
			}
		}
	}()

	return out
}

func logResults(ctx context.Context, results <-chan metricsUpdateResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-results:
			if !ok {
				return
			}
			if res.Err != nil {
				logger.Log.Errorf("Failed to update metric %s (%s): %v", res.Request.ID, res.Request.MType, res.Err)
			} else {
				logger.Log.Infof("Successfully updated metric %s (%s)", res.Request.ID, res.Request.MType)
			}
		}
	}
}
