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
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)

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

	alwaysValidValidator := func(metricType, metricName, metricValue string) error {
		return nil
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
				mockUpdater.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "update service error",
			url:  "/update/counter/requests/10",
			mockUpdate: func() {
				mockUpdater.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("update failed"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockUpdate()

			handler := MetricUpdatePathHandler(alwaysValidValidator, mockUpdater)
			req := makeRequest("POST", tt.url)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}
