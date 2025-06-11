package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/yp-metrics/internal/handlers"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
)

func TestMetricUpdateBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ptrFloat64 := func(f float64) *float64 {
		return &f
	}

	mockSvc := handlers.NewMockMetricUpdaterBody(ctrl)

	validMetric := types.Metrics{
		ID:    "testMetric",
		MType: types.Gauge,
		Value: ptrFloat64(123.45),
	}

	t.Run("success", func(t *testing.T) {
		mockSvc.EXPECT().Update(gomock.Any(), validMetric).Return(nil)

		validator := func(m types.Metrics) error {
			return nil
		}

		bodyBytes, _ := json.Marshal(validMetric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler := handlers.MetricUpdateBodyHandler(validator, mockSvc)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var respMetric types.Metrics
		err := json.NewDecoder(rec.Body).Decode(&respMetric)
		assert.NoError(t, err)
		assert.Equal(t, validMetric.ID, respMetric.ID)
		assert.Equal(t, validMetric.MType, respMetric.MType)
		assert.NotNil(t, respMetric.Value)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		validator := func(m types.Metrics) error { return nil }

		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte(`{"id":"testMetric", "type":"gauge",`))) // malformed JSON
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler := handlers.MetricUpdateBodyHandler(validator, mockSvc)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid JSON body")
	})

	t.Run("validation error - name missing", func(t *testing.T) {
		metric := types.Metrics{MType: types.Gauge, Value: ptrFloat64(1.23)}

		validator := func(m types.Metrics) error {
			return validators.ErrNameIsRequired
		}

		mockSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0) // should not call update on validation fail

		bodyBytes, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler := handlers.MetricUpdateBodyHandler(validator, mockSvc)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, validators.ErrNameIsRequired.Error()+"\n", rec.Body.String())
	})

	t.Run("service update error", func(t *testing.T) {
		validator := func(m types.Metrics) error { return nil }

		mockSvc.EXPECT().Update(gomock.Any(), validMetric).Return(errors.New("update error"))

		bodyBytes, _ := json.Marshal(validMetric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler := handlers.MetricUpdateBodyHandler(validator, mockSvc)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, types.ErrInternalServerError.Error()+"\n", rec.Body.String())
	})
}
