package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/yp-metrics/internal/handlers"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/middlewares"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/services"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
	"github.com/sbilibin2017/yp-metrics/internal/workers"
)

func main() {
	parseFlags()
	err := run(context.Background())
	if err != nil {
		panic(err)
	}
}

var (
	addr            string
	storeInterval   int
	fileStoragePath string
	restore         bool
	logLevel        string = "info"
)

func parseFlags() {
	flag.StringVar(&addr, "a", ":8080", "address and port to run server")
	flag.IntVar(&storeInterval, "i", 300, "store interval in seconds (0 = synchronous write)")
	flag.StringVar(&fileStoragePath, "f", "./data/metrics.json", "file storage path")
	flag.BoolVar(&restore, "r", true, "restore metrics from file at startup")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		if v, err := strconv.Atoi(envInterval); err == nil {
			storeInterval = v
		}
	}

	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		fileStoragePath = envPath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if v, err := strconv.ParseBool(envRestore); err == nil {
			restore = v
		}
	}
}

func run(ctx context.Context) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	data := make(map[types.MetricID]types.Metrics)

	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(data)
	metricFileSaveRepository := repositories.NewMetricFileSaveRepository(fileStoragePath)
	metricMemoryGetRepository := repositories.NewMetricMemoryGetRepository(data)
	metricMemoryListRepository := repositories.NewMetricMemoryListRepository(data)
	metricFileListRepository := repositories.NewMetricFileListRepository(fileStoragePath)

	metricUpdateService := services.NewMetricUpdateService(metricMemorySaveRepository, metricMemoryGetRepository)
	metricGetService := services.NewMetricGetService(metricMemoryGetRepository)
	metricListService := services.NewMetricListService(metricMemoryListRepository)

	metricUpdatePathHandler := handlers.MetricUpdatePathHandler(validators.ValidateMetricPath, metricUpdateService)
	metricUpdateBodyHandler := handlers.MetricUpdateBodyHandler(validators.ValidateMetricBody, metricUpdateService)
	metricGetPathHandler := handlers.MetricGetPathHandler(validators.ValidateMetricIDPath, metricGetService)
	metricGetBodyHandler := handlers.MetricGetBodyHandler(validators.ValidateMetricIDPath, metricGetService)
	metricListHTMLHandler := handlers.MetricListHTMLHandler(metricListService)

	middlewaresList := []func(http.Handler) http.Handler{
		middlewares.LoggingMiddleware,
		middlewares.GzipMiddleware,
	}

	router := chi.NewRouter()
	router.Use(middlewaresList...)

	router.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	router.Post("/update/{type}/{name}", metricUpdatePathHandler)
	router.Post("/update/", metricUpdateBodyHandler)

	router.Get("/value/{type}/{name}", metricGetPathHandler)
	router.Get("/value/{type}", metricGetPathHandler)
	router.Post("/value/", metricGetBodyHandler)

	router.Get("/", metricListHTMLHandler)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	go workers.StartMetricServerWorker(
		ctx,
		metricMemorySaveRepository,
		metricFileSaveRepository,
		metricMemoryListRepository,
		metricFileListRepository,
		storeInterval,
		restore,
	)

	errChan := make(chan error, 1)

	go func() {
		logger.Log.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, context.Canceled) && err != http.ErrServerClosed {
			logger.Log.Errorw("Server failed", "error", err)
			errChan <- err
		} else {
			logger.Log.Info("Server exited gracefully")
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("Shutdown signal received, shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Errorw("Server shutdown error", "error", err)
			return err
		}

		logger.Log.Info("Server shutdown completed gracefully")
		return ctx.Err()

	case err := <-errChan:
		if err != nil {
			logger.Log.Errorw("Server exited with error", "error", err)
		}
		return err
	}
}
