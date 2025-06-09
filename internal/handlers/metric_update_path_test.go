package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	gomock "github.com/golang/mock/gomock"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricFromPath(t *testing.T) {
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	tests := []struct {
		name         string
		metricType   string
		metricName   string
		metricValue  string
		wantMetric   *types.Metrics
		wantAPIError *types.APIError
	}{
		{
			name:         "valid gauge metric",
			metricType:   types.Gauge,
			metricName:   "temperature",
			metricValue:  "23.5",
			wantMetric:   &types.Metrics{ID: "temperature", MType: types.Gauge, Value: float64Ptr(23.5)},
			wantAPIError: nil,
		},
		{
			name:         "valid counter metric",
			metricType:   types.Counter,
			metricName:   "requests",
			metricValue:  "42",
			wantMetric:   &types.Metrics{ID: "requests", MType: types.Counter, Delta: int64Ptr(42)},
			wantAPIError: nil,
		},
		{
			name:         "missing metric name",
			metricType:   types.Gauge,
			metricName:   "",
			metricValue:  "10",
			wantMetric:   nil,
			wantAPIError: &types.APIError{StatusCode: http.StatusNotFound, Message: "Metric name is required"},
		},
		{
			name:         "missing metric type",
			metricType:   "",
			metricName:   "temperature",
			metricValue:  "10",
			wantMetric:   nil,
			wantAPIError: &types.APIError{StatusCode: http.StatusBadRequest, Message: "Metric type is required"},
		},
		{
			name:         "invalid metric type",
			metricType:   "foo",
			metricName:   "temperature",
			metricValue:  "10",
			wantMetric:   nil,
			wantAPIError: &types.APIError{StatusCode: http.StatusBadRequest, Message: "Invalid metric type"},
		},
		{
			name:         "missing metric value",
			metricType:   types.Gauge,
			metricName:   "temperature",
			metricValue:  "",
			wantMetric:   nil,
			wantAPIError: &types.APIError{StatusCode: http.StatusBadRequest, Message: "Metric value is required"},
		},
		{
			name:         "invalid gauge value",
			metricType:   types.Gauge,
			metricName:   "temperature",
			metricValue:  "abc",
			wantMetric:   nil,
			wantAPIError: &types.APIError{StatusCode: http.StatusBadRequest, Message: "Invalid gauge metric value"},
		},
		{
			name:         "invalid counter value",
			metricType:   types.Counter,
			metricName:   "requests",
			metricValue:  "abc",
			wantMetric:   nil,
			wantAPIError: &types.APIError{StatusCode: http.StatusBadRequest, Message: "Invalid counter metric value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetric, gotErr := newMetricFromPath(tt.metricType, tt.metricName, tt.metricValue)

			if tt.wantAPIError == nil {
				assert.Nil(t, gotErr)
				assert.NotNil(t, gotMetric)
				assert.Equal(t, tt.wantMetric.ID, gotMetric.ID)
				assert.Equal(t, tt.wantMetric.MType, gotMetric.MType)

				if gotMetric.MType == types.Gauge {
					assert.NotNil(t, gotMetric.Value)
					assert.InDelta(t, *tt.wantMetric.Value, *gotMetric.Value, 0.0001)
				} else {
					assert.NotNil(t, gotMetric.Delta)
					assert.Equal(t, *tt.wantMetric.Delta, *gotMetric.Delta)
				}
			} else {
				assert.NotNil(t, gotErr)
				assert.Nil(t, gotMetric)
				assert.Equal(t, tt.wantAPIError.StatusCode, gotErr.StatusCode)
				assert.Equal(t, tt.wantAPIError.Message, gotErr.Message)
			}
		})
	}
}

func TestMetricUpdateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)

	// Helper to create request with chi context including URL params
	makeRequest := func(method, url string) *http.Request {
		req := httptest.NewRequest(method, url, nil)
		rctx := chi.NewRouteContext()

		parts := strings.Split(url, "/")
		if len(parts) > 2 {
			rctx.URLParams.Add("type", parts[2])
		}
		if len(parts) > 3 {
			rctx.URLParams.Add("name", parts[3])
		}
		if len(parts) > 4 {
			rctx.URLParams.Add("value", parts[4])
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		return req
	}

	tests := []struct {
		name           string
		url            string
		mockUpdate     func()
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "valid gauge metric update",
			url:  "/update/gauge/temperature/23.5",
			mockUpdate: func() {
				val := 23.5
				mockUpdater.EXPECT().
					Update(gomock.Any(), gomock.Eq([]types.Metrics{
						{ID: "temperature", MType: types.Gauge, Value: &val},
					})).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
		{
			name: "missing metric name",
			url:  "/update/gauge//23.5",
			mockUpdate: func() {
				// Should not call Update since metric parsing fails
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "Metric name is required\n",
		},
		{
			name: "invalid metric type",
			url:  "/update/foo/temperature/23.5",
			mockUpdate: func() {
				// Should not call Update since metric parsing fails
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "Invalid metric type\n",
		},
		{
			name: "update service error",
			url:  "/update/counter/requests/10",
			mockUpdate: func() {
				val := int64(10)
				mockUpdater.EXPECT().
					Update(gomock.Any(), gomock.Eq([]types.Metrics{
						{ID: "requests", MType: types.Counter, Delta: &val},
					})).
					Return(errors.New("update failed"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "Metric is not updated\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockUpdate()

			handler := MetricUpdateHandler(mockUpdater)
			req := makeRequest("POST", tt.url)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}
