package validators

import (
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateMetricIDPath(t *testing.T) {
	tests := []struct {
		metricType string
		metricName string
		wantErr    error
	}{
		{"gauge", "cpu", nil},
		{"counter", "requests", nil},
		{"", "cpu", ErrTypeIsRequired},
		{"gauge", "", ErrNameIsRequired},
		{"invalid", "cpu", ErrInvalidMetricType},
	}

	for _, tt := range tests {
		err := ValidateMetricIDPath(tt.metricType, tt.metricName)
		assert.Equal(t, tt.wantErr, err)
	}
}

func TestValidateMetricPath(t *testing.T) {
	tests := []struct {
		metricType  string
		metricName  string
		metricValue string
		wantErr     error
	}{
		{"gauge", "cpu", "1.23", nil},
		{"counter", "requests", "10", nil},
		{"gauge", "cpu", "notafloat", ErrInvalidGaugeValue},
		{"counter", "requests", "notanint", ErrInvalidCounterValue},
		{"gauge", "", "1.23", ErrNameIsRequired},
		{"", "cpu", "1.23", ErrTypeIsRequired},
		{"gauge", "cpu", "", ErrValueIsRequired},
		{"invalid", "cpu", "1.23", ErrInvalidMetricType},
	}

	for _, tt := range tests {
		err := ValidateMetricPath(tt.metricType, tt.metricName, tt.metricValue)
		assert.Equal(t, tt.wantErr, err)
	}
}

func TestValidateMetricBody(t *testing.T) {
	v := float64(1.23)
	d := int64(10)

	tests := []struct {
		m       types.Metrics
		wantErr error
	}{
		{types.Metrics{ID: "cpu", MType: types.Gauge, Value: &v}, nil},
		{types.Metrics{ID: "req", MType: types.Counter, Delta: &d}, nil},
		{types.Metrics{ID: "", MType: types.Gauge, Value: &v}, ErrNameIsRequired},
		{types.Metrics{ID: "cpu", MType: "invalid", Value: &v}, ErrInvalidMetricType},
		{types.Metrics{ID: "cpu", MType: types.Gauge, Value: nil}, ErrValueIsRequired},
		{types.Metrics{ID: "req", MType: types.Counter, Delta: nil}, ErrValueIsRequired},
	}

	for _, tt := range tests {
		err := ValidateMetricBody(tt.m)
		assert.Equal(t, tt.wantErr, err)
	}
}
