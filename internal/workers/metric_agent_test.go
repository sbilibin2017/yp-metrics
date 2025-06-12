package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectRuntimeGaugeMetrics(t *testing.T) {
	metrics := collectRuntimeGaugeMetrics()
	require.NotEmpty(t, metrics, "metrics slice should not be empty")

	expectedIDs := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys",
		"LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse",
		"MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
		"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue",
	}

	foundIDs := make(map[string]bool)
	for _, m := range metrics {
		require.NotEmpty(t, m.ID)
		require.Contains(t, expectedIDs, m.ID)
		require.Equal(t, types.Gauge, m.MType)

		require.NotNil(t, m.Value)

		foundIDs[m.ID] = true
	}

	for _, id := range expectedIDs {
		require.Contains(t, foundIDs, id, "metric %s not found", id)
	}
}

func TestCollectRuntimeCounterMetrics(t *testing.T) {
	metrics := collectRuntimeCounterMetrics()
	require.Len(t, metrics, 1, "expected exactly one metric")

	m := metrics[0]
	require.Equal(t, "PollCount", m.ID)
	require.Equal(t, types.Counter, m.MType)

}

func TestPollMetrics(t *testing.T) {
	collector := func() []types.Metrics {
		return []types.Metrics{
			{ID: "testMetric", MType: types.Gauge, Value: float64PtrToStringPtr(123.456)},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pollInterval := 1

	outCh := pollMetrics(ctx, pollInterval, []func() []types.Metrics{collector})

	select {
	case metricsBatch := <-outCh:
		require.Len(t, metricsBatch, 1)
		m := metricsBatch[0]
		require.Equal(t, "testMetric", m.ID)
		require.Equal(t, types.Gauge, m.MType)
		require.NotNil(t, m.Value)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for metric from pollMetrics")
	}

	cancel()

	// Wait for the channel to close and verify it's closed
	select {
	case _, ok := <-outCh:
		require.False(t, ok, "expected channel to be closed after context cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for channel to close after context cancel")
	}
}
func float64PtrToStringPtr(f float64) *float64 {
	return &f
}

func TestReportMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricsUpdater(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reportInterval := 1

	in := make(chan []types.Metrics)

	firstMetric := types.Metrics{ID: "metric1", MType: types.Gauge, Value: float64PtrToStringPtr(1)}
	secondMetric := types.Metrics{ID: "metric2", MType: types.Counter, Delta: int64Ptr(2)}

	// Expect one call with both metrics combined
	mockUpdater.EXPECT().Updates(gomock.Any(), []types.Metrics{firstMetric, secondMetric}).Return(nil).Times(1)

	outCh := reportMetrics(ctx, mockUpdater, reportInterval, in)

	in <- []types.Metrics{firstMetric}
	in <- []types.Metrics{secondMetric}
	close(in)

	time.Sleep(1100 * time.Millisecond)
	cancel()

	var results []metricsUpdateResult
	for res := range outCh {
		results = append(results, res)
	}

	counts := map[string]int{}
	for _, r := range results {
		require.NoError(t, r.Err)
		for _, m := range r.Request {
			counts[m.ID]++
		}
	}

	require.Equal(t, 1, counts["metric1"])
	require.Equal(t, 1, counts["metric2"])
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestReportMetrics_TickerCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricsUpdater(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reportInterval := 1
	in := make(chan []types.Metrics)

	metrics := []types.Metrics{
		{ID: "metric1", MType: types.Gauge, Value: float64PtrToStringPtr(1)},
		{ID: "metric2", MType: types.Counter, Delta: int64Ptr(2)},
	}

	// Expect a single batch update call with all metrics
	mockUpdater.EXPECT().Updates(gomock.Any(), metrics).Return(nil).Times(1)

	outCh := reportMetrics(ctx, mockUpdater, reportInterval, in)

	// Send metrics as a batch
	in <- metrics
	close(in) // CLOSE input channel so reportMetrics knows no more data coming

	// Wait a bit more than reportInterval to allow ticker to trigger
	time.Sleep(time.Duration(reportInterval)*time.Second + 200*time.Millisecond)

	results := make(map[string]bool)

	// Correct loop to break properly on channel close or when results are complete
	for {
		select {
		case res, ok := <-outCh:
			if !ok {
				goto doneReading // channel closed, exit loop
			}
			require.NoError(t, res.Err)
			for _, m := range res.Request {
				results[m.ID] = true
			}
			if len(results) == len(metrics) {
				goto doneReading // got all metrics, exit loop early
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for metrics update result")
		}
	}

doneReading:

	for _, m := range metrics {
		assert.True(t, results[m.ID], "metric %s not processed", m.ID)
	}

	cancel()

	// Drain channel to avoid goroutine leaks (optional here, channel should be closed)
	for range outCh {
	}
}

func TestLogResults(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan metricsUpdateResult)

	done := make(chan struct{})
	go func() {
		logResults(ctx, results)
		close(done)
	}()

	results <- metricsUpdateResult{
		Request: []types.Metrics{
			{
				ID:    "metric_success",
				MType: types.Gauge,
				Value: float64PtrToStringPtr(100),
			},
		},
		Err: nil,
	}

	results <- metricsUpdateResult{
		Request: []types.Metrics{
			{
				ID:    "metric_fail",
				MType: types.Counter,
				Delta: int64Ptr(200),
			},
		},
		Err: errors.New("some error"),
	}

	close(results)

	<-done
}

func TestLogResults_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan metricsUpdateResult)

	done := make(chan struct{})
	go func() {
		logResults(ctx, results)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("logResults did not stop after context was canceled")
	}
}

func TestStartMetricAgentWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricsUpdater(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockUpdater.
		EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(nil).
		MinTimes(1)

	pollInterval := 1
	reportInterval := 1

	go StartMetricAgentWorker(ctx, mockUpdater, pollInterval, reportInterval)

	time.Sleep(1500 * time.Millisecond)

	cancel()

	time.Sleep(500 * time.Millisecond)
}
