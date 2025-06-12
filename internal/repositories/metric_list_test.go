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

func TestMetricListerContext_List_NoStrategy(t *testing.T) {
	lister := repositories.NewMetricListerContext()

	metrics, err := lister.List(context.Background())

	require.Nil(t, metrics)
	require.Error(t, err)
	require.Equal(t, "strategy is not set", err.Error())
}

func TestMetricListerContext_List_WithStrategy_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := repositories.NewMockLister(ctrl)

	expectedMetrics := []types.Metrics{
		{ID: "metric1", MType: "gauge"},
		{ID: "metric2", MType: "counter"},
	}

	mockLister.EXPECT().
		List(gomock.Any()).
		Return(expectedMetrics, nil).
		Times(1)

	lister := repositories.NewMetricListerContext()
	lister.SetContext(mockLister)

	metrics, err := lister.List(context.Background())

	require.NoError(t, err)
	require.Equal(t, expectedMetrics, metrics)
}

func TestMetricListerContext_List_WithStrategy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := repositories.NewMockLister(ctrl)

	expectedErr := errors.New("list failed")

	mockLister.EXPECT().
		List(gomock.Any()).
		Return(nil, expectedErr).
		Times(1)

	lister := repositories.NewMetricListerContext()
	lister.SetContext(mockLister)

	metrics, err := lister.List(context.Background())

	require.Nil(t, metrics)
	require.Equal(t, expectedErr, err)
}
