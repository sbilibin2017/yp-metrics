package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
		expected    *Metrics
		expectedErr error
	}{
		{
			name:        "valid gauge",
			metricType:  Gauge,
			metricName:  "temperature",
			metricValue: "42.42",
			expected: &Metrics{
				ID:    "temperature",
				MType: Gauge,
				Value: float64Ptr(42.42),
			},
			expectedErr: nil,
		},
		{
			name:        "valid counter",
			metricType:  Counter,
			metricName:  "requests",
			metricValue: "10",
			expected: &Metrics{
				ID:    "requests",
				MType: Counter,
				Delta: int64Ptr(10),
			},
			expectedErr: nil,
		},
		{
			name:        "missing name",
			metricType:  Gauge,
			metricName:  "",
			metricValue: "10.0",
			expected:    nil,
			expectedErr: ErrNameIsRequired,
		},
		{
			name:        "missing type",
			metricType:  "",
			metricName:  "test",
			metricValue: "10.0",
			expected:    nil,
			expectedErr: ErrTypeIsRequired,
		},
		{
			name:        "invalid type",
			metricType:  "foo",
			metricName:  "test",
			metricValue: "10.0",
			expected:    nil,
			expectedErr: ErrInvalidMetricType,
		},
		{
			name:        "missing value",
			metricType:  Gauge,
			metricName:  "test",
			metricValue: "",
			expected:    nil,
			expectedErr: ErrValueIsRequired,
		},
		{
			name:        "invalid gauge value",
			metricType:  Gauge,
			metricName:  "test",
			metricValue: "abc",
			expected:    nil,
			expectedErr: ErrInvalidGaugeValue,
		},
		{
			name:        "invalid counter value",
			metricType:  Counter,
			metricName:  "test",
			metricValue: "abc",
			expected:    nil,
			expectedErr: ErrInvalidCounterValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewMetrics(tt.metricType, tt.metricName, tt.metricValue)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.MType, result.MType)
				if tt.expected.MType == Gauge {
					assert.NotNil(t, result.Value)
					assert.InDelta(t, *tt.expected.Value, *result.Value, 0.0001)
				}
				if tt.expected.MType == Counter {
					assert.NotNil(t, result.Delta)
					assert.Equal(t, *tt.expected.Delta, *result.Delta)
				}
			} else {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			}
		})
	}
}
