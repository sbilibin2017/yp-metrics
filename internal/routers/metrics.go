package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMetricsRouter(
	metricUpdatePathHandler http.HandlerFunc,
	metricGetPathHandler http.HandlerFunc,
	metricListHTMLHandler http.HandlerFunc,
) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	router.Post("/update/{type}/{name}", metricUpdatePathHandler)
	router.Get("/value/{type}/{name}", metricGetPathHandler)
	router.Get("/value/{type}", metricGetPathHandler)
	router.Get("/", metricListHTMLHandler)
	return router
}
