package handlers

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"

	"github.com/go-chi/chi/v5"
)

type MetricUpdater interface {
	Update(ctx context.Context, metrics types.Metrics) error
}

func MetricUpdatePathHandler(
	val func(metricType string, metricName string, metricValue string) error,
	svc MetricUpdater,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		err := val(metricType, metricName, metricValue)

		if err != nil {
			switch err {
			case validators.ErrNameIsRequired:
				http.Error(w, err.Error(), http.StatusNotFound)
			case validators.ErrInvalidMetricType,
				validators.ErrInvalidGaugeValue,
				validators.ErrInvalidCounterValue,
				validators.ErrTypeIsRequired,
				validators.ErrValueIsRequired:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return

		}

		metric := types.NewMetrics(metricType, metricName, metricValue)

		if err := svc.Update(r.Context(), *metric); err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
