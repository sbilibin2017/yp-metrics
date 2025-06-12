package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetPathHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockMetricGetterPath(ctrl)

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
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		return req
	}

	alwaysValidValidator := func(metricType, metricName string) error {
		return nil
	}

	tests := []struct {
		name           string
		url            string
		mockGetter     func()
		wantStatusCode int
		wantBody       string
		validator      func(string, string) error
	}{
		{
			name: "valid gauge metric",
			url:  "/value/gauge/load",
			mockGetter: func() {
				val := 42.5
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "load", MType: types.Gauge}).
					Return(&types.Metrics{ID: "load", MType: types.Gauge, Value: &val}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "42.5",
			validator:      alwaysValidValidator,
		},
		{
			name: "valid counter metric",
			url:  "/value/counter/hits",
			mockGetter: func() {
				val := int64(100)
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "hits", MType: types.Counter}).
					Return(&types.Metrics{ID: "hits", MType: types.Counter, Delta: &val}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "100",
			validator:      alwaysValidValidator,
		},
		{
			name: "metric not found",
			url:  "/value/gauge/unknown",
			mockGetter: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "unknown", MType: types.Gauge}).
					Return(nil, types.ErrMetricNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       types.ErrMetricNotFound.Error() + "\n",
			validator:      alwaysValidValidator,
		},
		{
			name: "internal service error",
			url:  "/value/gauge/internal",
			mockGetter: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "internal", MType: types.Gauge}).
					Return(nil, errors.New("fail"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       types.ErrInternalServerError.Error() + "\n",
			validator:      alwaysValidValidator,
		},
		{
			name:           "invalid type",
			url:            "/value/invalidtype/load",
			mockGetter:     func() {}, // no call expected
			wantStatusCode: http.StatusBadRequest,
			wantBody:       validators.ErrInvalidMetricType.Error() + "\n",
			validator: func(metricType, metricName string) error {
				return validators.ErrInvalidMetricType
			},
		},
		{
			name:           "missing name",
			url:            "/value/gauge/",
			mockGetter:     func() {}, // no call expected
			wantStatusCode: http.StatusNotFound,
			wantBody:       validators.ErrNameIsRequired.Error() + "\n",
			validator: func(metricType, metricName string) error {
				return validators.ErrNameIsRequired
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockGetter()

			handler := MetricGetPathHandler(tt.validator, mockGetter)

			req := makeRequest("GET", tt.url)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}
