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
	"github.com/stretchr/testify/assert"
)

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
	}{
		{
			name: "valid gauge metric update",
			url:  "/update/gauge/temperature/23.5",
			mockUpdate: func() {
				val := 23.5
				mockUpdater.EXPECT().
					Update(gomock.Any(), types.Metrics{
						ID:    "temperature",
						MType: types.Gauge,
						Value: &val,
					}).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "missing metric name",
			url:  "/update/gauge//23.5",
			mockUpdate: func() {
				// Should not call Update
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "invalid metric type",
			url:  "/update/foo/temperature/23.5",
			mockUpdate: func() {
				// Should not call Update
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "update service error",
			url:  "/update/counter/requests/10",
			mockUpdate: func() {
				val := int64(10)
				mockUpdater.EXPECT().
					Update(gomock.Any(), types.Metrics{
						ID:    "requests",
						MType: types.Counter,
						Delta: &val,
					}).
					Return(errors.New("update failed"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockUpdate()

			handler := MetricUpdatePathHandler(mockUpdater)
			req := makeRequest("POST", tt.url)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}
