package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

func TestMetricUpdateService_Update(t *testing.T) {
	ptrInt64 := func(i int64) *int64 {
		return &i
	}

	tests := []struct {
		name     string
		input    []types.Metrics
		expected []types.Metrics
		setup    func(ctrl *gomock.Controller) MetricSaver
		wantErr  bool
	}{
		{
			name: "overwrite gauge metric",
			input: []types.Metrics{
				{ID: "foo", MType: types.Gauge, Delta: ptrInt64(10)},
				{ID: "foo", MType: types.Gauge, Delta: ptrInt64(20)},
			},
			expected: []types.Metrics{
				{ID: "foo", MType: types.Gauge, Delta: ptrInt64(20)}, // overwritten by last
			},
			setup: func(ctrl *gomock.Controller) MetricSaver {
				mockSaver := NewMockMetricSaver(ctrl)
				mockSaver.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, got []types.Metrics) error {
						assert.ElementsMatch(t, []types.Metrics{
							{ID: "foo", MType: types.Gauge, Delta: ptrInt64(20)},
						}, got)
						return nil
					})
				return mockSaver
			},
		},
		{
			name: "sum counter metrics",
			input: []types.Metrics{
				{ID: "bar", MType: types.Counter, Delta: ptrInt64(2)},
				{ID: "bar", MType: types.Counter, Delta: ptrInt64(3)},
			},
			expected: []types.Metrics{
				{ID: "bar", MType: types.Counter, Delta: ptrInt64(5)}, // sum 2 + 3
			},
			setup: func(ctrl *gomock.Controller) MetricSaver {
				mockSaver := NewMockMetricSaver(ctrl)
				mockSaver.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, got []types.Metrics) error {
						assert.ElementsMatch(t, []types.Metrics{
							{ID: "bar", MType: types.Counter, Delta: ptrInt64(5)},
						}, got)
						return nil
					})
				return mockSaver
			},
		},
		{
			name: "save error propagates",
			input: []types.Metrics{
				{ID: "fail", MType: types.Counter, Delta: ptrInt64(1)},
			},
			expected: nil,
			setup: func(ctrl *gomock.Controller) MetricSaver {
				mockSaver := NewMockMetricSaver(ctrl)
				mockSaver.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(errors.New("save failed"))
				return mockSaver
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSaver := tt.setup(ctrl)
			svc := NewMetricUpdateService(mockSaver)

			err := svc.Update(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
