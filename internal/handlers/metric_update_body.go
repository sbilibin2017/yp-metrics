package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
)

type MetricUpdaterBody interface {
	Update(ctx context.Context, metric types.Metrics) error
}

func MetricUpdateBodyHandler(
	val func(m types.Metrics) error,
	svc MetricUpdaterBody,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric types.Metrics

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&metric); err != nil {
			http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

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

		if err := svc.Update(r.Context(), metric); err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metric); err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}
	}
}
