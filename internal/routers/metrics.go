package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMetricsRouter(
	metricUpdatePathHandler http.HandlerFunc,
	metricUpdateBodyHandler http.HandlerFunc,
	metricGetPathHandler http.HandlerFunc,
	metricGetBodyandler http.HandlerFunc,
	metricListHTMLHandler http.HandlerFunc,
	middlewares ...func(next http.Handler) http.Handler,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middlewares...)

	router.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	router.Post("/update/{type}/{name}", metricUpdatePathHandler)
	router.Post("/update/", metricUpdateBodyHandler)

	router.Get("/value/{type}/{name}", metricGetPathHandler)
	router.Get("/value/{type}", metricGetPathHandler)
	router.Post("/value/", metricGetBodyandler)

	router.Get("/", metricListHTMLHandler)

	return router
}
