package handlers

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricLister interface {
	List(ctx context.Context) ([]types.Metrics, error)
}

func MetricListHTMLHandler(svc MetricLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsList, err := svc.List(r.Context())
		if err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}

		html, err := types.GetMetricsHTML(metricsList)
		if err != nil {
			http.Error(w, types.ErrInternalServerError.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}
}
