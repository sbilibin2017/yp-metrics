package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

func TestLoadMetricsFromFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileLister := NewMockMetricsFileLister(ctrl)
	mockMemorySaver := NewMockMetricsMemorySaver(ctrl)

	expectedMetrics := []types.Metrics{
		{ID: "metric1", MType: "counter", Delta: ptrInt64(5)},
		{ID: "metric2", MType: "gauge", Value: ptrFloat64(3.14)},
	}

	mockFileLister.EXPECT().List(gomock.Any()).Return(expectedMetrics, nil).Times(1)
	for _, m := range expectedMetrics {
		mockMemorySaver.EXPECT().Save(gomock.Any(), m).Return(nil)
	}

	loadMetricsFromFile(context.Background(), mockFileLister, mockMemorySaver)
}

func TestLoadMetricsFromFile_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileLister := NewMockMetricsFileLister(ctrl)
	mockMemorySaver := NewMockMetricsMemorySaver(ctrl)

	someErr := errors.New("list error")
	mockFileLister.EXPECT().List(gomock.Any()).Return(nil, someErr).Times(1)

	loadMetricsFromFile(context.Background(), mockFileLister, mockMemorySaver)
}

func TestSaveAllMetrics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemoryLister := NewMockMetricsMemoryLister(ctrl)
	mockFileSaver := NewMockMetricsFileSaver(ctrl)

	expectedMetrics := []types.Metrics{
		{ID: "metric1", MType: "counter", Delta: ptrInt64(1)},
		{ID: "metric2", MType: "gauge", Value: ptrFloat64(2.71)},
	}

	mockMemoryLister.EXPECT().List(gomock.Any()).Return(expectedMetrics, nil).Times(1)
	for _, m := range expectedMetrics {
		mockFileSaver.EXPECT().Save(gomock.Any(), m).Return(nil)
	}

	saveMetricsToFile(context.Background(), mockMemoryLister, mockFileSaver)
}

func TestSaveAllMetrics_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemoryLister := NewMockMetricsMemoryLister(ctrl)
	mockFileSaver := NewMockMetricsFileSaver(ctrl)

	someErr := errors.New("list error")
	mockMemoryLister.EXPECT().List(gomock.Any()).Return(nil, someErr).Times(1)

	saveMetricsToFile(context.Background(), mockMemoryLister, mockFileSaver)
}

func ptrInt64(i int64) *int64 {
	return &i
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func TestMetricServerWorker_RestoreAndInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ms := NewMockMetricsMemorySaver(ctrl)
	fs := NewMockMetricsFileSaver(ctrl)
	ml := NewMockMetricsMemoryLister(ctrl)
	fl := NewMockMetricsFileLister(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// mock restore
	mockMetric := types.Metrics{ID: "test", MType: "gauge", Value: ptrFloat64(3.14)}
	fl.EXPECT().List(gomock.Any()).Return([]types.Metrics{mockMetric}, nil)
	ms.EXPECT().Save(gomock.Any(), mockMetric).Return(nil)

	// mock periodic save
	ml.EXPECT().List(gomock.Any()).Return([]types.Metrics{mockMetric}, nil).AnyTimes()
	fs.EXPECT().Save(gomock.Any(), mockMetric).Return(nil).AnyTimes()

	go func() {

		StartMetricServerWorker(ctx, ms, fs, ml, fl, 1, true)
	}()

	time.Sleep(1500 * time.Millisecond)
	cancel()
}

func TestMetricServerWorker_StoreIntervalZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ms := NewMockMetricsMemorySaver(ctrl)
	fs := NewMockMetricsFileSaver(ctrl)
	ml := NewMockMetricsMemoryLister(ctrl)
	fl := NewMockMetricsFileLister(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockMetric := types.Metrics{ID: "cpu", MType: "counter", Delta: ptrInt64(42)}
	ml.EXPECT().List(gomock.Any()).Return([]types.Metrics{mockMetric}, nil)
	fs.EXPECT().Save(gomock.Any(), mockMetric).Return(nil)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	StartMetricServerWorker(ctx, ms, fs, ml, fl, 0, false)
}
