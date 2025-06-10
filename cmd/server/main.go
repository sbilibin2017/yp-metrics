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

var (
	logLevel   string = "info"
	runAddress string = ":8080"
)

func main() {
	config := configs.NewServerConfig(
		configs.WithServerLogLevel(logLevel),
		configs.WithServerRunAddress(runAddress),
	)

	if err := run(context.Background(), config); err != nil {
		panic(err)
	}
}

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
	metricRepo := repositories.NewMetricMemorySaveRepository(data)

	metricService := services.NewMetricUpdateService(metricRepo)
	metricHandler := handlers.MetricUpdateHandler(metricService)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", metricHandler)
	router.Post("/update/{type}/{name}", metricHandler)

	return &http.Server{
		Addr:    config.RunAddress,
		Handler: router,
	}, nil

}
