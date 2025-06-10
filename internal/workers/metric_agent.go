package workers

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type metricsUpdateResult struct {
	Request types.MetricsUpdatePathRequest
	Err     error
}

type MetricsUpdater interface {
	Update(ctx context.Context, req types.MetricsUpdatePathRequest) error
}

func NewMetricAgentWorker(
	metricsUpdater MetricsUpdater,
	pollInterval int,
	reportInterval int,
) func(ctx context.Context) {
	return func(ctx context.Context) {
		pollMetricsCh := pollMetrics(
			ctx,
			[]func() []types.MetricsUpdatePathRequest{collectRuntimeCounterMetrics, collectRuntimeGaugeMetrics},
			pollInterval,
		)
		reportMetricsCh := reportMetrics(
			ctx,
			metricsUpdater,
			reportInterval,
			pollMetricsCh,
		)
		logResults(ctx, reportMetricsCh)
	}
}

func pollMetrics(
	ctx context.Context,
	collectors []func() []types.MetricsUpdatePathRequest,
	pollInterval int,
) <-chan types.MetricsUpdatePathRequest {

	out := make(chan types.MetricsUpdatePathRequest)
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)

	fanIn := func(chans []<-chan types.MetricsUpdatePathRequest) <-chan types.MetricsUpdatePathRequest {
		var wg sync.WaitGroup
		merged := make(chan types.MetricsUpdatePathRequest)

		output := func(c <-chan types.MetricsUpdatePathRequest) {
			defer wg.Done()
			for m := range c {
				merged <- m
			}
		}

		wg.Add(len(chans))
		for _, c := range chans {
			go output(c)
		}

		go func() {
			wg.Wait()
			close(merged)
		}()

		return merged
	}

	fanOutCollector := func(
		ctx context.Context,
		collector func() []types.MetricsUpdatePathRequest,
	) []<-chan types.MetricsUpdatePathRequest {
		chs := make([]<-chan types.MetricsUpdatePathRequest, len(collectors))

		for i := 0; i < len(collectors); i++ {
			ch := make(chan types.MetricsUpdatePathRequest)
			chs[i] = ch

			go func(ch chan<- types.MetricsUpdatePathRequest) {
				defer close(ch)
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						metrics := collector()
						for _, m := range metrics {
							ch <- m
						}
					}
				}
			}(ch)
		}
		return chs
	}

	go func() {
		defer ticker.Stop()
		defer close(out)

		var allChans []<-chan types.MetricsUpdatePathRequest
		for _, collector := range collectors {
			chs := fanOutCollector(ctx, collector)
			allChans = append(allChans, chs...)
		}

		mergedCh := fanIn(allChans)

		for m := range mergedCh {
			out <- m
		}
	}()

	return out
}

func collectRuntimeGaugeMetrics() []types.MetricsUpdatePathRequest {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := []types.MetricsUpdatePathRequest{
		{ID: "Alloc", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.Alloc))},
		{ID: "BuckHashSys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.BuckHashSys))},
		{ID: "Frees", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.Frees))},
		{ID: "GCCPUFraction", MType: types.Gauge, Value: fmt.Sprintf("%f", memStats.GCCPUFraction)},
		{ID: "GCSys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.GCSys))},
		{ID: "HeapAlloc", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.HeapAlloc))},
		{ID: "HeapIdle", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.HeapIdle))},
		{ID: "HeapInuse", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.HeapInuse))},
		{ID: "HeapObjects", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.HeapObjects))},
		{ID: "HeapReleased", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.HeapReleased))},
		{ID: "HeapSys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.HeapSys))},
		{ID: "LastGC", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.LastGC))},
		{ID: "Lookups", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.Lookups))},
		{ID: "MCacheInuse", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.MCacheInuse))},
		{ID: "MCacheSys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.MCacheSys))},
		{ID: "MSpanInuse", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.MSpanInuse))},
		{ID: "MSpanSys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.MSpanSys))},
		{ID: "Mallocs", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.Mallocs))},
		{ID: "NextGC", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.NextGC))},
		{ID: "NumForcedGC", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.NumForcedGC))},
		{ID: "NumGC", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.NumGC))},
		{ID: "OtherSys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.OtherSys))},
		{ID: "PauseTotalNs", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.PauseTotalNs))},
		{ID: "StackInuse", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.StackInuse))},
		{ID: "StackSys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.StackSys))},
		{ID: "Sys", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.Sys))},
		{ID: "TotalAlloc", MType: types.Gauge, Value: fmt.Sprintf("%f", float64(memStats.TotalAlloc))},
		{ID: "RandomValue", MType: types.Gauge, Value: fmt.Sprintf("%f", rand.Float64()*100)},
	}

	return metrics
}

func collectRuntimeCounterMetrics() []types.MetricsUpdatePathRequest {
	metrics := []types.MetricsUpdatePathRequest{
		{ID: "PollCount", MType: types.Counter, Value: fmt.Sprintf("%d", 1)},
	}
	return metrics
}

func reportMetrics(
	ctx context.Context,
	metricsUpdater MetricsUpdater,
	reportInterval int,
	in <-chan types.MetricsUpdatePathRequest,
) <-chan metricsUpdateResult {
	out := make(chan metricsUpdateResult)
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	go func() {
		defer func() {
			ticker.Stop()
			close(out)
		}()

		var buffer []types.MetricsUpdatePathRequest

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
	go func() {
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
					logger.Log.Infof("Successfully updated metric %s (%s): %s", res.Request.ID, res.Request.MType, res.Request.Value)
				}
			}
		}
	}()
}
