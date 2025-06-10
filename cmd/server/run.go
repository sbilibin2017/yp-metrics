package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/handlers"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/runners"
	"github.com/sbilibin2017/yp-metrics/internal/services"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

func run(ctx context.Context, config *configs.ServerConfig) error {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	ctx, stop := runners.NewRunContext(ctx)
	defer stop()

	srv, err := newServer(config)
	if err != nil {
		return err
	}

	err = runners.RunServer(ctx, srv)
	if err != nil && err != context.Canceled {
		return err
	}

	return nil
}

func newServer(config *configs.ServerConfig) (*http.Server, error) {
	data := make(map[types.MetricID]types.Metrics)

	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(data)
	metricMemoryGetRepository := repositories.NewMetricMemoryGetRepository(data)
	metricMemoryListRepository := repositories.NewMetricMemoryListRepository(data)

	metricUpdateService := services.NewMetricUpdateService(metricMemorySaveRepository, metricMemoryGetRepository)
	metricGetService := services.NewMetricGetService(metricMemoryGetRepository)
	metricListService := services.NewMetricListService(metricMemoryListRepository)

	metricUpdatePathHandler := handlers.MetricUpdatePathHandler(metricUpdateService)
	metricGetPathHandler := handlers.MetricGetPathHandler(metricGetService)
	metricListHTMLHandler := handlers.MetricListHTMLHandler(metricListService)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	router.Post("/update/{type}/{name}", metricUpdatePathHandler)

	router.Get("/value/{type}/{name}", metricGetPathHandler)
	router.Get("/value/{type}", metricGetPathHandler)

	router.Get("/", metricListHTMLHandler)

	return &http.Server{
		Addr:    config.RunAddress,
		Handler: router,
	}, nil

}
