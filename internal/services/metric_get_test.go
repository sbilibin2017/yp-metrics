package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockMetricGetter(ctrl)
	svc := NewMetricGetService(mockGetter)

	ctx := context.Background()
	metricID := types.MetricID{ID: "testMetric", MType: types.Gauge}
	expectedMetric := &types.Metrics{
		ID:    "testMetric",
		MType: types.Gauge,
		Value: func() *float64 { v := 123.45; return &v }(),
	}

	// Тест успешного вызова
	mockGetter.EXPECT().
		Get(ctx, metricID).
		Return(expectedMetric, nil).
		Times(1)

	result, err := svc.Get(ctx, metricID)
	assert.NoError(t, err)
	assert.Equal(t, expectedMetric, result)

	// Тест случая ошибки
	expectedErr := errors.New("not found")
	mockGetter.EXPECT().
		Get(ctx, metricID).
		Return(nil, expectedErr).
		Times(1)

	result, err = svc.Get(ctx, metricID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}
