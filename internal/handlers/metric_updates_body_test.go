package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
	"github.com/stretchr/testify/assert"
)

// Define a simple validator func matching your signature
func alwaysValid(_ types.Metrics) error { return nil }
func alwaysInvalid(err error) func(types.Metrics) error {
	return func(_ types.Metrics) error { return err }
}

func TestMetricUpdatesBodyHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricUpdatersBody(ctrl)

	metrics := []types.Metrics{
		{ID: "metric1", MType: types.Gauge, Value: func(v float64) *float64 { return &v }(123.45)},
		{ID: "metric2", MType: types.Counter, Delta: func(v int64) *int64 { return &v }(10)},
	}

	// Expect Update to be called for each metric and return nil error
	for _, m := range metrics {
		mockSvc.EXPECT().Update(gomock.Any(), m).Return(nil)
	}

	bodyBytes, _ := json.Marshal(metrics)
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	handler := MetricUpdatesBodyHandler(alwaysValid, mockSvc)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// The response body should echo back the metrics JSON
	var got []types.Metrics
	err := json.NewDecoder(rec.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, metrics, got)
}

func TestMetricUpdatesBodyHandler_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricUpdatersBody(ctrl)

	badJSON := `[{id: "missing quotes"}]`
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte(badJSON)))
	rec := httptest.NewRecorder()

	handler := MetricUpdatesBodyHandler(alwaysValid, mockSvc)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid JSON body")
}

func TestMetricUpdatesBodyHandler_ValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricUpdatersBody(ctrl)

	metrics := []types.Metrics{{ID: "", MType: types.Gauge, Value: func(v float64) *float64 { return &v }(1)}}
	bodyBytes, _ := json.Marshal(metrics)

	tests := []struct {
		validatorErr error
		wantStatus   int
	}{
		{validators.ErrNameIsRequired, http.StatusNotFound},
		{validators.ErrInvalidMetricType, http.StatusBadRequest},
		{validators.ErrInvalidGaugeValue, http.StatusBadRequest},
		{validators.ErrInvalidCounterValue, http.StatusBadRequest},
		{validators.ErrTypeIsRequired, http.StatusBadRequest},
		{validators.ErrValueIsRequired, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.validatorErr.Error(), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler := MetricUpdatesBodyHandler(alwaysInvalid(tt.validatorErr), mockSvc)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.validatorErr.Error())
		})
	}
}

func TestMetricUpdatesBodyHandler_UpdateServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricUpdatersBody(ctrl)

	metrics := []types.Metrics{{ID: "m1", MType: types.Gauge, Value: func(v float64) *float64 { return &v }(3.14)}}
	bodyBytes, _ := json.Marshal(metrics)

	mockSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	handler := MetricUpdatesBodyHandler(alwaysValid, mockSvc)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), types.ErrInternalServerError.Error())
}
