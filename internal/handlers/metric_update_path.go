package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sbilibin2017/yp-metrics/internal/types"

	"github.com/go-chi/chi/v5"
)

// MetricUpdater defines the interface for updating metrics.
type MetricUpdater interface {
	// Update processes a slice of metrics and persists changes.
	Update(ctx context.Context, metrics []types.Metrics) error
}

// MetricUpdateHandler returns an HTTP handler function that processes
// metric update requests with metric type, name, and value passed as URL parameters.
//
// The handler extracts parameters, validates and parses them into a Metrics struct,
// then invokes the provided MetricUpdater service to update the metrics.
// If validation or update fails, appropriate HTTP error responses are returned.
//
// Expected URL parameters:
//   - type: metric type (gauge or counter)
//   - name: metric identifier string
//   - value: metric value (string representation of float64 for gauge, int64 for counter)
func MetricUpdateHandler(svc MetricUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		metric, apiErr := newMetricFromPath(metricType, metricName, metricValue)
		if apiErr != nil {
			http.Error(w, apiErr.Message, apiErr.StatusCode)
			return
		}

		err := svc.Update(r.Context(), []types.Metrics{*metric})
		if err != nil {
			http.Error(w, "Metric is not updated", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// newMetricFromPath validates metric parameters and creates a new Metrics instance.
//
// Parameters:
//   - metricType: the metric type string ("gauge" or "counter")
//   - metricName: the metric identifier string (non-empty)
//   - metricValue: the metric value string representation
//
// Returns:
//   - A pointer to a populated Metrics struct if validation and parsing succeed.
//   - An APIError pointer if any validation or parsing error occurs.
func newMetricFromPath(metricType string, metricName string, metricValue string) (*types.Metrics, *types.APIError) {
	if metricName == "" {
		return nil, &types.APIError{StatusCode: http.StatusNotFound, Message: "Metric name is required"}
	}

	if metricType == "" {
		return nil, &types.APIError{StatusCode: http.StatusBadRequest, Message: "Metric type is required"}
	}

	if metricType != types.Gauge && metricType != types.Counter {
		return nil, &types.APIError{StatusCode: http.StatusBadRequest, Message: "Invalid metric type"}
	}

	if metricValue == "" {
		return nil, &types.APIError{StatusCode: http.StatusBadRequest, Message: "Metric value is required"}
	}

	metric := &types.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case types.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return nil, &types.APIError{StatusCode: http.StatusBadRequest, Message: "Invalid gauge metric value"}
		}
		metric.Value = &val

	case types.Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return nil, &types.APIError{StatusCode: http.StatusBadRequest, Message: "Invalid counter metric value"}
		}
		metric.Delta = &val
	}

	return metric, nil
}
