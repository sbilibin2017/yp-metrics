package validators

import (
	"errors"
	"strconv"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

var (
	ErrNameIsRequired      = errors.New("metric name is required")
	ErrTypeIsRequired      = errors.New("metric type is required")
	ErrInvalidMetricType   = errors.New("invalid metric type")
	ErrValueIsRequired     = errors.New("metric value is required")
	ErrInvalidGaugeValue   = errors.New("invalid gauge metric value")
	ErrInvalidCounterValue = errors.New("invalid counter metric value")
)

func ValidateMetricIDPath(metricType, metricName string) error {
	if metricName == "" {
		return ErrNameIsRequired
	}
	if metricType == "" {
		return ErrTypeIsRequired
	}
	if metricType != types.Gauge && metricType != types.Counter {
		return ErrInvalidMetricType
	}
	return nil
}

func ValidateMetricPath(metricType, metricName, metricValue string) error {
	if metricName == "" {
		return ErrNameIsRequired
	}
	if metricType == "" {
		return ErrTypeIsRequired
	}
	if metricType != types.Gauge && metricType != types.Counter {
		return ErrInvalidMetricType
	}
	if metricValue == "" {
		return ErrValueIsRequired
	}

	switch metricType {
	case types.Gauge:
		if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
			return ErrInvalidGaugeValue
		}
	case types.Counter:
		if _, err := strconv.ParseInt(metricValue, 10, 64); err != nil {
			return ErrInvalidCounterValue
		}
	}
	return nil
}

func ValidateMetricBody(m types.Metrics) error {
	if m.ID == "" {
		return ErrNameIsRequired
	}
	if m.MType != types.Gauge && m.MType != types.Counter {
		return ErrInvalidMetricType
	}
	switch m.MType {
	case types.Gauge:
		if m.Value == nil {
			return ErrValueIsRequired
		}
	case types.Counter:
		if m.Delta == nil {
			return ErrValueIsRequired
		}
	}
	return nil
}
