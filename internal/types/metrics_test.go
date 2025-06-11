package types_test

import (
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	gauge := types.NewMetrics(types.Gauge, "testGauge", "12.34")
	assert.Equal(t, "testGauge", gauge.ID)
	assert.Equal(t, types.Gauge, gauge.MType)
	assert.NotNil(t, gauge.Value)
	assert.Nil(t, gauge.Delta)
	assert.Equal(t, 12.34, *gauge.Value)

	counter := types.NewMetrics(types.Counter, "testCounter", "42")
	assert.Equal(t, "testCounter", counter.ID)
	assert.Equal(t, types.Counter, counter.MType)
	assert.NotNil(t, counter.Delta)
	assert.Nil(t, counter.Value)
	assert.Equal(t, int64(42), *counter.Delta)

	// If metricValue invalid, NewMetrics returns zero value (no error handling here)
	invalidGauge := types.NewMetrics(types.Gauge, "badGauge", "abc")
	assert.NotNil(t, invalidGauge)
	assert.Nil(t, invalidGauge.Value) // parsing failed

	invalidCounter := types.NewMetrics(types.Counter, "badCounter", "abc")
	assert.NotNil(t, invalidCounter)
	assert.Nil(t, invalidCounter.Delta) // parsing failed
}

func TestNewMetricID(t *testing.T) {
	mType := "gauge"
	id := "metric1"

	metricID := types.NewMetricID(mType, id)

	assert.NotNil(t, metricID)
	assert.Equal(t, mType, metricID.MType)
	assert.Equal(t, id, metricID.ID)
}

func TestGetMetricValueString(t *testing.T) {
	gaugeVal := 3.14
	counterVal := int64(10)

	tests := []struct {
		name    string
		metric  types.Metrics
		want    string
		wantErr error
	}{
		{
			name: "valid gauge",
			metric: types.Metrics{
				MType: types.Gauge,
				Value: &gaugeVal,
			},
			want:    "3.14",
			wantErr: nil,
		},
		{
			name: "nil gauge value",
			metric: types.Metrics{
				MType: types.Gauge,
				Value: nil,
			},
			want:    "",
			wantErr: types.ErrNilMetricValue,
		},
		{
			name: "valid counter",
			metric: types.Metrics{
				MType: types.Counter,
				Delta: &counterVal,
			},
			want:    "10",
			wantErr: nil,
		},
		{
			name: "nil counter value",
			metric: types.Metrics{
				MType: types.Counter,
				Delta: nil,
			},
			want:    "",
			wantErr: types.ErrNilMetricValue,
		},
		{
			name: "unknown type",
			metric: types.Metrics{
				MType: "unknown",
			},
			want:    "",
			wantErr: types.ErrUnknownMType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.GetMetricValueString(tt.metric)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMetricsHTML(t *testing.T) {
	gaugeVal := 1.23
	counterVal := int64(5)

	metrics := []types.Metrics{
		{
			ID:    "metric1",
			MType: types.Gauge,
			Value: &gaugeVal,
		},
		{
			ID:    "metric2",
			MType: types.Counter,
			Delta: &counterVal,
		},
		{
			ID:    "metric3",
			MType: types.Gauge,
			Value: nil, // should be skipped
		},
	}

	html, err := types.GetMetricsHTML(metrics)
	assert.NoError(t, err)
	assert.Contains(t, html, "<li>metric1 (gauge): 1.23</li>")
	assert.Contains(t, html, "<li>metric2 (counter): 5</li>")
	assert.NotContains(t, html, "metric3") // nil value skipped
}
