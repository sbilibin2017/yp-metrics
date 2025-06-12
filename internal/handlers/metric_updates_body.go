package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
)

type MetricUpdatersBody interface {
	Update(ctx context.Context, metric types.Metrics) error
}

func MetricUpdatesBodyHandler(
	val func(m types.Metrics) error,
	svc MetricUpdatersBody,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []types.Metrics

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&metrics); err != nil {
			http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

		for _, metric := range metrics {
			err := val(metric)

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
		}

		for _, metric := range metrics {
			if err := svc.Update(r.Context(), metric); err != nil {
				http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}
	}
}
