package handlers

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/yp-metrics/internal/types"

	"github.com/go-chi/chi/v5"
)

type MetricGetterPath interface {
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

func MetricGetPathHandler(svc MetricGetterPath) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		metricID, err := types.NewMetricID(metricType, metricName)
		if err != nil {
			switch err {
			case types.ErrNameIsRequired:
				http.Error(w, err.Error(), http.StatusNotFound)
			case types.ErrInvalidMetricType,
				types.ErrInvalidGaugeValue,
				types.ErrInvalidCounterValue,
				types.ErrTypeIsRequired:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		metric, err := svc.Get(r.Context(), *metricID)

		if err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}

		if metric == nil {
			http.Error(w, types.ErrMetricNotFound.Error(), http.StatusNotFound)
			return

		}

		valueString, err := types.GetMetricValueString(*metric)

		if err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return

		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(valueString))
	}
}
