package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/require"
)

func TestMetricGetterContext_Get_NoStrategy(t *testing.T) {
	getter := repositories.NewMetricGetterContext()

	metric, err := getter.Get(context.Background(), types.MetricID{ID: "test", MType: "gauge"})

	require.Nil(t, metric)
	require.Error(t, err)
	require.Equal(t, "strategy is not set", err.Error())
}

func TestMetricGetterContext_Get_WithStrategy_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := repositories.NewMockGetter(ctrl)

	id := types.MetricID{ID: "metric1", MType: "gauge"}
	expectedMetric := &types.Metrics{
		ID:    "metric1",
		MType: "gauge",
		// add other fields if needed
	}

	mockGetter.EXPECT().
		Get(gomock.Any(), id).
		Return(expectedMetric, nil).
		Times(1)

	getter := repositories.NewMetricGetterContext()
	getter.SetContext(mockGetter)

	metric, err := getter.Get(context.Background(), id)

	require.NoError(t, err)
	require.Equal(t, expectedMetric, metric)
}

func TestMetricGetterContext_Get_WithStrategy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := repositories.NewMockGetter(ctrl)

	id := types.MetricID{ID: "metric2", MType: "counter"}
	expectedErr := errors.New("get failed")

	mockGetter.EXPECT().
		Get(gomock.Any(), id).
		Return(nil, expectedErr).
		Times(1)

	getter := repositories.NewMetricGetterContext()
	getter.SetContext(mockGetter)

	metric, err := getter.Get(context.Background(), id)

	require.Nil(t, metric)
	require.Equal(t, expectedErr, err)
}
