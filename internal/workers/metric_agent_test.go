package workers

import (
	"context"
	"errors"
	"strconv"
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

		require.NotEmpty(t, m.Value)
		require.True(t, isFloatString(m.Value), "value for metric %s should be a float string", m.ID)

		foundIDs[m.ID] = true
	}

	for _, id := range expectedIDs {
		require.Contains(t, foundIDs, id, "metric %s not found", id)
	}
}

func isFloatString(s string) bool {
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c == '.' || c == '-' || c == 'e' || c == 'E') {
			return false
		}
	}
	return true
}

func TestCollectRuntimeCounterMetrics(t *testing.T) {
	metrics := collectRuntimeCounterMetrics()
	require.Len(t, metrics, 1, "expected exactly one metric")

	m := metrics[0]
	require.Equal(t, "PollCount", m.ID)
	require.Equal(t, types.Counter, m.MType)

	val, err := strconv.ParseInt(m.Value, 10, 64)
	require.NoError(t, err, "value should be parseable as int64")
	require.Equal(t, int64(1), val, "expected PollCount value to be 1")
}

func TestPollMetrics(t *testing.T) {
	collector := func() []types.MetricsUpdatePathRequest {
		return []types.MetricsUpdatePathRequest{
			{ID: "testMetric", MType: types.Gauge, Value: "123.456"},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pollInterval := 1

	outCh := pollMetrics(ctx, []func() []types.MetricsUpdatePathRequest{collector}, pollInterval)

	select {
	case m := <-outCh:
		require.Equal(t, "testMetric", m.ID)
		require.Equal(t, types.Gauge, m.MType)
		require.Equal(t, "123.456", m.Value)

	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for metric from pollMetrics")
	}

	cancel()

	_, ok := <-outCh
	require.False(t, ok, "expected channel to be closed after context cancel")
}

func TestReportMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricsUpdater(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reportInterval := 1

	in := make(chan types.MetricsUpdatePathRequest)

	firstMetric := types.MetricsUpdatePathRequest{ID: "metric1", MType: types.Gauge, Value: "1"}
	secondMetric := types.MetricsUpdatePathRequest{ID: "metric2", MType: types.Counter, Value: "2"}

	mockUpdater.EXPECT().Update(gomock.Any(), firstMetric).Return(nil).Times(1)
	mockUpdater.EXPECT().Update(gomock.Any(), secondMetric).Return(nil).Times(1)

	outCh := reportMetrics(ctx, mockUpdater, reportInterval, in)

	in <- firstMetric
	in <- secondMetric

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
		counts[r.Request.ID]++
	}

	require.Equal(t, 1, counts["metric1"])
	require.Equal(t, 1, counts["metric2"])
}

func TestReportMetrics_TickerCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricsUpdater(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reportInterval := 1

	in := make(chan types.MetricsUpdatePathRequest)

	metrics := []types.MetricsUpdatePathRequest{
		{ID: "metric1", MType: types.Gauge, Value: "1"},
		{ID: "metric2", MType: types.Counter, Value: "2"},
	}

	for _, m := range metrics {
		mockUpdater.EXPECT().Update(gomock.Any(), m).Return(nil).Times(1)
	}

	outCh := reportMetrics(ctx, mockUpdater, reportInterval, in)

	for _, m := range metrics {
		in <- m
	}

	time.Sleep(time.Duration(reportInterval)*time.Second + 200*time.Millisecond)

	results := make(map[string]bool)
	for i := 0; i < len(metrics); i++ {
		select {
		case res := <-outCh:
			require.NoError(t, res.Err)
			results[res.Request.ID] = true
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for metrics update result")
		}
	}

	for _, m := range metrics {
		assert.True(t, results[m.ID], "metric %s not processed", m.ID)
	}

	close(in)
	cancel()

	for range outCh {
	}
}

func TestReportMetrics_ContextDone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricsUpdater(ctrl)
	ctx, cancel := context.WithCancel(context.Background())

	reportInterval := 10

	in := make(chan types.MetricsUpdatePathRequest)

	metrics := []types.MetricsUpdatePathRequest{
		{ID: "metric1", MType: types.Gauge, Value: "1"},
		{ID: "metric2", MType: types.Counter, Value: "2"},
	}

	for _, m := range metrics {
		mockUpdater.EXPECT().Update(gomock.Any(), m).Return(nil).Times(1)
	}

	outCh := reportMetrics(ctx, mockUpdater, reportInterval, in)

	for _, m := range metrics {
		in <- m
	}

	cancel()

	results := make(map[string]bool)
	for i := 0; i < len(metrics); i++ {
		select {
		case res := <-outCh:
			require.NoError(t, res.Err)
			results[res.Request.ID] = true
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for metrics update result")
		}
	}

	for _, m := range metrics {
		assert.True(t, results[m.ID], "metric %s not processed", m.ID)
	}

	close(in)

	_, ok := <-outCh
	assert.False(t, ok, "output channel should be closed")
}

func TestLogResults(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan metricsUpdateResult)
	logResults(ctx, results)
	results <- metricsUpdateResult{
		Request: types.MetricsUpdatePathRequest{
			ID:    "metric_success",
			MType: types.Gauge,
			Value: "100",
		},
		Err: nil,
	}
	results <- metricsUpdateResult{
		Request: types.MetricsUpdatePathRequest{
			ID:    "metric_fail",
			MType: types.Counter,
			Value: "200",
		},
		Err: errors.New("some error"),
	}
	close(results)
	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestLogResults_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan metricsUpdateResult)
	logResults(ctx, results)
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestStartMetricAgentWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricsUpdater(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockUpdater.
		EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil).
		MinTimes(1)

	pollInterval := 1
	reportInterval := 1

	w := NewMetricAgentWorker(mockUpdater, pollInterval, reportInterval)

	go w(ctx)

	time.Sleep(1500 * time.Millisecond)

	cancel()

	time.Sleep(500 * time.Millisecond)
}
