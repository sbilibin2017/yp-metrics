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

func TestMetricSaver_Save_NoStrategy(t *testing.T) {
	ms := repositories.NewMetricSaverContext()

	err := ms.Save(context.Background(), types.Metrics{})
	require.Error(t, err)
	require.Equal(t, "strategy is not set", err.Error())
}

func TestMetricSaver_Save_WithStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := repositories.NewMockSaver(ctrl)

	metric := types.Metrics{
		ID:    "metric1",
		MType: "gauge",
		// other fields...
	}

	ms := &repositories.MetricSaverContext{}
	ms.SetContext(mockSaver)

	// Expect Save to be called once with context and metric, and return nil error
	mockSaver.EXPECT().
		Save(gomock.Any(), metric).
		Return(nil).
		Times(1)

	err := ms.Save(context.Background(), metric)
	require.NoError(t, err)
}

func TestMetricSaver_Save_WithStrategy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := repositories.NewMockSaver(ctrl)

	metric := types.Metrics{
		ID:    "metric2",
		MType: "counter",
	}

	ms := &repositories.MetricSaverContext{}
	ms.SetContext(mockSaver)

	expectedErr := errors.New("save failed")

	mockSaver.EXPECT().
		Save(gomock.Any(), metric).
		Return(expectedErr).
		Times(1)

	err := ms.Save(context.Background(), metric)
	require.Equal(t, expectedErr, err)
}
