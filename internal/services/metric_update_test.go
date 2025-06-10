package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/yp-metrics/internal/services"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateService_Update(t *testing.T) {
	type fields struct {
		setupMocks func(*services.MockMetricUpdateSaver, *services.MockMetricUpdateGetter)
	}
	type args struct {
		metrics types.Metrics
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     error
		expectedVal int64
	}{
		{
			name: "counter metric - existing value added",
			fields: fields{
				setupMocks: func(saver *services.MockMetricUpdateSaver, getter *services.MockMetricUpdateGetter) {
					getter.EXPECT().Get(gomock.Any(), types.MetricID{ID: "requests", MType: types.Counter}).
						Return(&types.Metrics{ID: "requests", MType: types.Counter, Delta: int64Ptr(5)}, nil)
					saver.EXPECT().Save(gomock.Any(), types.Metrics{ID: "requests", MType: types.Counter, Delta: int64Ptr(15)}).
						Return(nil)
				},
			},
			args: args{
				metrics: types.Metrics{ID: "requests", MType: types.Counter, Delta: int64Ptr(10)},
			},
			wantErr:     nil,
			expectedVal: 15,
		},
		{
			name: "getter fails",
			fields: fields{
				setupMocks: func(saver *services.MockMetricUpdateSaver, getter *services.MockMetricUpdateGetter) {
					getter.EXPECT().Get(gomock.Any(), types.MetricID{ID: "fail_metric", MType: types.Counter}).
						Return(nil, errors.New("db error"))
				},
			},
			args: args{
				metrics: types.Metrics{ID: "fail_metric", MType: types.Counter, Delta: int64Ptr(2)},
			},
			wantErr: types.ErrInternalServerError,
		},
		{
			name: "gauge metric - saved directly",
			fields: fields{
				setupMocks: func(saver *services.MockMetricUpdateSaver, getter *services.MockMetricUpdateGetter) {
					saver.EXPECT().Save(gomock.Any(), types.Metrics{ID: "temp", MType: types.Gauge, Value: float64Ptr(42.42)}).
						Return(nil)
				},
			},
			args: args{
				metrics: types.Metrics{ID: "temp", MType: types.Gauge, Value: float64Ptr(42.42)},
			},
			wantErr:     nil,
			expectedVal: 0,
		},
		{
			name: "save fails",
			fields: fields{
				setupMocks: func(saver *services.MockMetricUpdateSaver, getter *services.MockMetricUpdateGetter) {
					saver.EXPECT().Save(gomock.Any(), gomock.Any()).
						Return(errors.New("save error"))
				},
			},
			args: args{
				metrics: types.Metrics{ID: "some", MType: types.Gauge, Value: float64Ptr(100)},
			},
			wantErr: errors.New("save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSaver := services.NewMockMetricUpdateSaver(ctrl)
			mockGetter := services.NewMockMetricUpdateGetter(ctrl)
			tt.fields.setupMocks(mockSaver, mockGetter)

			svc := services.NewMetricUpdateService(mockSaver, mockGetter)
			err := svc.Update(context.Background(), tt.args.metrics)

			assert.Equal(t, tt.wantErr, err)
			if tt.args.metrics.MType == types.Counter && tt.wantErr == nil {
				assert.Equal(t, tt.expectedVal, *tt.args.metrics.Delta)
			}
		})
	}
}

// helpers
func int64Ptr(v int64) *int64 {
	return &v
}
func float64Ptr(v float64) *float64 {
	return &v
}
