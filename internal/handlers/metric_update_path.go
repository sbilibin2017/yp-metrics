package handlers

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/yp-metrics/internal/types"

	"github.com/go-chi/chi/v5"
)

type MetricUpdater interface {
	Update(ctx context.Context, metrics []types.Metrics) error
}

func MetricUpdateHandler(svc MetricUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		metric, err := types.NewMetrics(metricType, metricName, metricValue)
		if err != nil {
			switch err {
			case types.ErrNameIsRequired:
				http.Error(w, err.Error(), http.StatusNotFound)
			case types.ErrInvalidMetricType,
				types.ErrInvalidGaugeValue,
				types.ErrInvalidCounterValue,
				types.ErrTypeIsRequired,
				types.ErrValueIsRequired:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		if err := svc.Update(r.Context(), []types.Metrics{*metric}); err != nil {
			http.Error(w, "metric is not updated", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
