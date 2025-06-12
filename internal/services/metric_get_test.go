package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetService_Get_TableDriven(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockMetricGetter(ctrl)
	svc := NewMetricGetService(mockGetter)

	ctx := context.Background()

	expectedMetric := &types.Metrics{
		ID:    "testMetric",
		MType: types.Gauge,
		Value: func() *float64 { v := 123.45; return &v }(),
	}

	tests := []struct {
		name         string
		metricID     types.MetricID
		returnMetric *types.Metrics
		returnErr    error
		expectedErr  error
	}{
		{
			name:         "successful retrieval",
			metricID:     types.MetricID{ID: "testMetric", MType: types.Gauge},
			returnMetric: expectedMetric,
			returnErr:    nil,
			expectedErr:  nil,
		},
		{
			name:         "getter returns error",
			metricID:     types.MetricID{ID: "testMetric", MType: types.Gauge},
			returnMetric: nil,
			returnErr:    errors.New("some error"),
			expectedErr:  types.ErrInternalServerError,
		},
		{
			name:         "metric not found (nil metric, no error)",
			metricID:     types.MetricID{ID: "missingMetric", MType: types.Gauge},
			returnMetric: nil,
			returnErr:    nil,
			expectedErr:  types.ErrMetricNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter.EXPECT().
				Get(ctx, tt.metricID).
				Return(tt.returnMetric, tt.returnErr).
				Times(1)

			result, err := svc.Get(ctx, tt.metricID)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.returnMetric, result)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
