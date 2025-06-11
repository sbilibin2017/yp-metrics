package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
}

func NewMetrics(metricType string, metricName string, metricValue string) *Metrics {
	metric := &Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			metric.Value = nil
		} else {
			metric.Value = &val
		}

	case Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			metric.Delta = nil
		} else {
			metric.Delta = &val
		}
	}

	return metric
}

type MetricID struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}

var (
	ErrMetricNotFound = errors.New("metric not found")
)

func NewMetricID(mType string, id string) *MetricID {
	return &MetricID{ID: id, MType: mType}
}

var (
	ErrNilMetricValue = errors.New("metric value is nil")
	ErrUnknownMType   = errors.New("unknown metric type")
)

func GetMetricValueString(metric Metrics) (string, error) {
	switch metric.MType {
	case Counter:
		if metric.Delta == nil {
			return "", ErrNilMetricValue
		}
		return strconv.FormatInt(*metric.Delta, 10), nil
	case Gauge:
		if metric.Value == nil {
			return "", ErrNilMetricValue
		}
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64), nil
	default:
		return "", ErrUnknownMType
	}
}

var (
	ErrInternalServerError = errors.New("internal server error")
)

func GetMetricsHTML(metricsList []Metrics) (string, error) {
	var builder strings.Builder

	builder.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	builder.WriteString("<meta charset=\"UTF-8\">\n<title>Metrics</title>\n")
	builder.WriteString("</head>\n<body>\n<h1>Metrics</h1>\n<ul>\n")

	for _, metric := range metricsList {
		valueStr, err := GetMetricValueString(metric)
		if err != nil {
			continue
		}
		line := fmt.Sprintf("<li>%s (%s): %s</li>\n", metric.ID, metric.MType, valueStr)
		builder.WriteString(line)
	}

	builder.WriteString("</ul>\n</body>\n</html>")
	return builder.String(), nil
}
