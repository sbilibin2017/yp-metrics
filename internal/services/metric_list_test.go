package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricListService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := NewMockMetricLister(ctrl)
	svc := NewMetricListService(mockLister)

	ctx := context.Background()
	expectedMetrics := []types.Metrics{
		{
			ID:    "metric1",
			MType: types.Gauge,
			Value: func() *float64 { v := 10.5; return &v }(),
		},
		{
			ID:    "metric2",
			MType: types.Counter,
			Delta: func() *int64 { v := int64(5); return &v }(),
		},
	}

	// Тест успешного вызова
	mockLister.EXPECT().
		List(ctx).
		Return(expectedMetrics, nil).
		Times(1)

	result, err := svc.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedMetrics, result)

	// Тест случая ошибки
	expectedErr := errors.New("list error")
	mockLister.EXPECT().
		List(ctx).
		Return(nil, expectedErr).
		Times(1)

	result, err = svc.List(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}
