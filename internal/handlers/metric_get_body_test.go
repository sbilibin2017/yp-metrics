package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
)

func TestMetricGetBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	floatPtr := func(f float64) *float64 {
		return &f
	}

	mockSvc := NewMockMetricGetterBody(ctrl)

	validate := func(metricType, metricName string) error {
		if metricName == "" {
			return validators.ErrNameIsRequired
		}
		if metricType != "gauge" && metricType != "counter" {
			return validators.ErrInvalidMetricType
		}
		return nil
	}

	handler := MetricGetBodyHandler(validate, mockSvc)

	t.Run("valid request returns metric JSON", func(t *testing.T) {
		metricID := types.MetricID{ID: "metric1", MType: "gauge"}
		expectedMetric := &types.Metrics{
			ID:    "metric1",
			MType: "gauge",
			Value: floatPtr(123.45),
		}

		// Setup mock expectation
		mockSvc.EXPECT().
			Get(gomock.Any(), metricID).
			Return(expectedMetric, nil).
			Times(1)

		bodyBytes, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/some-url", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gotMetric types.Metrics
		err := json.NewDecoder(resp.Body).Decode(&gotMetric)
		assert.NoError(t, err)
		assert.Equal(t, *expectedMetric, gotMetric)
	})

	t.Run("invalid JSON returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/some-url", bytes.NewReader([]byte(`{invalid json`)))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error returns proper status", func(t *testing.T) {
		testCases := []struct {
			name         string
			metricID     types.MetricID
			expectedCode int
		}{
			{"missing name", types.MetricID{ID: "", MType: "gauge"}, http.StatusNotFound},
			{"invalid type", types.MetricID{ID: "metric1", MType: "invalid"}, http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				bodyBytes, _ := json.Marshal(tc.metricID)
				req := httptest.NewRequest(http.MethodPost, "/some-url", bytes.NewReader(bodyBytes))
				w := httptest.NewRecorder()

				handler(w, req)

				resp := w.Result()
				defer resp.Body.Close()

				assert.Equal(t, tc.expectedCode, resp.StatusCode)
			})
		}
	})

	t.Run("service returns error results in 500", func(t *testing.T) {
		metricID := types.MetricID{ID: "metric1", MType: "gauge"}

		mockSvc.EXPECT().
			Get(gomock.Any(), metricID).
			Return(nil, errors.New("some error")).
			Times(1)

		bodyBytes, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/some-url", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("metric not found returns 404", func(t *testing.T) {
		metricID := types.MetricID{ID: "metric1", MType: "gauge"}

		mockSvc.EXPECT().
			Get(gomock.Any(), metricID).
			Return(nil, nil).
			Times(1)

		bodyBytes, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/some-url", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
