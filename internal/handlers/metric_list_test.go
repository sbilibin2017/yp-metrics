package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricListHTMLHandler(t *testing.T) {
	float64Ptr := func(f float64) *float64 {
		return &f
	}

	int64Ptr := func(i int64) *int64 {
		return &i
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := NewMockMetricLister(ctrl)

	handler := MetricListHTMLHandler(mockLister)

	t.Run("success", func(t *testing.T) {
		metrics := []types.Metrics{
			{ID: "load", MType: types.Gauge, Value: float64Ptr(12.3)},
			{ID: "hits", MType: types.Counter, Delta: int64Ptr(42)},
		}

		mockLister.EXPECT().List(gomock.Any()).Return(metrics, nil)

		// Для теста подменим GetMetricsHTML, если нужно, но тут просто проверим, что вернется непустой HTML.
		// Можно подменить через функцию в types (если сделана как переменная) — если нет, используем настоящую.

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, rec.Body.String(), "load")
		assert.Contains(t, rec.Body.String(), "hits")
	})

	t.Run("service error", func(t *testing.T) {
		mockLister.EXPECT().List(gomock.Any()).Return(nil, errors.New("fail"))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), types.ErrInternalServerError.Error())
	})
}
