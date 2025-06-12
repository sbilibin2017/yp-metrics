package handlers

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"

	"github.com/go-chi/chi/v5"
)

type MetricGetterPath interface {
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

func MetricGetPathHandler(
	val func(metricType string, metricName string) error,
	svc MetricGetterPath,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		err := val(metricType, metricName)

		if err != nil {
			switch err {
			case validators.ErrNameIsRequired:
				http.Error(w, err.Error(), http.StatusNotFound)
			case validators.ErrInvalidMetricType,
				validators.ErrTypeIsRequired:
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return

		}

		metricID := types.NewMetricID(metricType, metricName)
		if metricID == nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}

		metric, err := svc.Get(r.Context(), *metricID)

		if err != nil {
			switch err {
			case types.ErrMetricNotFound:
				http.Error(w, types.ErrMetricNotFound.Error(), http.StatusNotFound)
				return
			default:
				http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
				return
			}
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
