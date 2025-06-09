package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/yp-metrics/internal/handlers"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/services"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

const (
	// logLevel defines the logging level for the application.
	logLevel = "info"

	// serverAddr is the address the HTTP server listens on.
	serverAddr = ":8080"
)

// main is the entry point of the application.
// It runs the HTTP server and panics if an error occurs.
func main() {
	err := run(context.Background())
	if err != nil {
		panic(err)
	}
}

// run initializes dependencies, sets up the HTTP server with routing,
// and starts serving requests until the provided context is canceled.
// It returns an error if initialization or server execution fails.
//
// The server exposes a POST endpoint to update metrics at:
//
//	/update/{type}/{name}/{value}
//
// The context controls the server lifecycle, allowing graceful shutdown.
func run(ctx context.Context) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}
	logger.Log.Info("Logger initialized")

	data := make(map[types.MetricID]types.Metrics)
	metricRepo := repositories.NewMetricMemorySaveRepository(data)

	metricService := services.NewMetricUpdateService(metricRepo)
	metricHandler := handlers.MetricUpdateHandler(metricService)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", metricHandler)
	router.Post("/update/{type}/{name}", metricHandler)

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	logger.Log.Infow("Starting HTTP server", "addr", serverAddr)

	err := runServer(ctx, srv)
	if err != nil && err != context.Canceled {
		logger.Log.Errorw("HTTP server stopped with error", "error", err)
		return err
	}

	logger.Log.Info("HTTP server stopped gracefully")
	return nil
}

func runServer(ctx context.Context, srv *http.Server) error {
	errChan := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, context.Canceled) && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}

		return ctx.Err()

	case err := <-errChan:
		return err
	}
}
