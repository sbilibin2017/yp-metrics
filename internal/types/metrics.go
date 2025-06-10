package types

import (
	"errors"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type MetricsUpdatePathRequest struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value string `json:"value"`
}

var (
	ErrNameIsRequired      = errors.New("metric name is required")
	ErrTypeIsRequired      = errors.New("metric type is required")
	ErrInvalidMetricType   = errors.New("invalid metric type")
	ErrValueIsRequired     = errors.New("metric value is required")
	ErrInvalidGaugeValue   = errors.New("invalid gauge metric value")
	ErrInvalidCounterValue = errors.New("invalid counter metric value")
)

func NewMetrics(metricType string, metricName string, metricValue string) (*Metrics, error) {
	if metricName == "" {
		return nil, ErrNameIsRequired
	}

	if metricType == "" {
		return nil, ErrTypeIsRequired
	}

	if metricType != Gauge && metricType != Counter {
		return nil, ErrInvalidMetricType
	}

	if metricValue == "" {
		return nil, ErrValueIsRequired
	}

	metric := &Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return nil, ErrInvalidGaugeValue
		}
		metric.Value = &val

	case Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return nil, ErrInvalidCounterValue
		}
		metric.Delta = &val
	}

	return metric, nil
}

type MetricID struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}
