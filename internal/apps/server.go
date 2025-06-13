package apps

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/contexts"
	"github.com/sbilibin2017/yp-metrics/internal/handlers"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/middlewares"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/services"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
	"github.com/sbilibin2017/yp-metrics/internal/workers"
)

type ServerApp struct {
	config  *configs.ServerConfig
	db      *sqlx.DB
	server  *http.Server
	workers []func(ctx context.Context)
}

func NewServerApp(config *configs.ServerConfig) (*ServerApp, error) {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return nil, err
	}

	var (
		db  *sqlx.DB
		err error
	)

	if config.DatabaseDSN != "" {
		db, err = newDB(config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
	}

	data := make(map[types.MetricID]types.Metrics)

	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(data)
	metricFileSaveRepository := repositories.NewMetricFileSaveRepository(config.FileStoragePath)
	metricDBSaveRepository := repositories.NewMetricDBSaveRepository(db, contexts.GetTxFromContext)

	metricMemoryGetRepository := repositories.NewMetricMemoryGetRepository(data)
	metricFileGetRepository := repositories.NewMetricFileGetRepository(config.FileStoragePath)
	metricDBGetRepository := repositories.NewMetricDBGetRepository(db, contexts.GetTxFromContext)

	metricMemoryListRepository := repositories.NewMetricMemoryListRepository(data)
	metricFileListRepository := repositories.NewMetricFileListRepository(config.FileStoragePath)
	metricDBListRepository := repositories.NewMetricDBListRepository(db, contexts.GetTxFromContext)

	metricSaverContext := repositories.NewMetricSaverContext()
	metricGetterContext := repositories.NewMetricGetterContext()
	metricListerContext := repositories.NewMetricListerContext()

	if config.DatabaseDSN != "" {
		metricSaverContext.SetContext(metricDBSaveRepository)
		metricGetterContext.SetContext(metricDBGetRepository)
		metricListerContext.SetContext(metricDBListRepository)
		logger.Log.Info("Using database repositories for saver, getter, and lister")
	} else if config.FileStoragePath != "" {
		metricSaverContext.SetContext(metricFileSaveRepository)
		metricGetterContext.SetContext(metricFileGetRepository)
		metricListerContext.SetContext(metricFileListRepository)
		logger.Log.Infow("Using file repositories for saver, getter, and lister", "fileStoragePath", config.FileStoragePath)
	} else {
		metricSaverContext.SetContext(metricMemorySaveRepository)
		metricGetterContext.SetContext(metricMemoryGetRepository)
		metricListerContext.SetContext(metricMemoryListRepository)
		logger.Log.Info("Using in-memory repositories for saver, getter, and lister")
	}

	metricUpdateService := services.NewMetricUpdateService(metricSaverContext, metricGetterContext)
	metricGetService := services.NewMetricGetService(metricGetterContext)
	metricListService := services.NewMetricListService(metricListerContext)

	logger.Log.Info("Services initialized")

	metricUpdatePathHandler := handlers.MetricUpdatePathHandler(validators.ValidateMetricPath, metricUpdateService)
	metricUpdateBodyHandler := handlers.MetricUpdateBodyHandler(validators.ValidateMetricBody, metricUpdateService)
	metricUpdatesBodyHandler := handlers.MetricUpdatesBodyHandler(validators.ValidateMetricBody, metricUpdateService)
	metricGetPathHandler := handlers.MetricGetPathHandler(validators.ValidateMetricIDPath, metricGetService)
	metricGetBodyHandler := handlers.MetricGetBodyHandler(validators.ValidateMetricIDPath, metricGetService)
	metricListHTMLHandler := handlers.MetricListHTMLHandler(metricListService)
	pingDBHandler := handlers.PingDBHandler(db)

	middlewares := []func(http.Handler) http.Handler{
		middlewares.LoggingMiddleware,
		middlewares.GzipMiddleware,
		middlewares.TxMiddleware(db, contexts.SetTxToContext),
		middlewares.RetryMiddleware,
		middlewares.HashMiddleware(config.HashHeader, config.HashKey),
	}

	router := chi.NewRouter()
	router.Use(middlewares...)

	router.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	router.Post("/update/{type}/{name}", metricUpdatePathHandler)
	router.Post("/update/", metricUpdateBodyHandler)
	router.Post("/updates/", metricUpdatesBodyHandler)

	router.Get("/value/{type}/{name}", metricGetPathHandler)
	router.Get("/value/{type}", metricGetPathHandler)
	router.Post("/value/", metricGetBodyHandler)

	router.Get("/", metricListHTMLHandler)

	router.Get("/ping", pingDBHandler)

	srv := &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	ws := make([]func(ctx context.Context), 0)
	if config.FileStoragePath != "" {
		ws = append(ws, func(ctx context.Context) {
			workers.StartMetricServerWorker(
				ctx,
				metricMemorySaveRepository,
				metricFileSaveRepository,
				metricMemoryListRepository,
				metricFileListRepository,
				config.StoreInterval,
				config.Restore,
			)
		})
	}

	app := &ServerApp{
		config:  config,
		db:      db,
		server:  srv,
		workers: ws,
	}

	return app, nil
}

func (a *ServerApp) Start(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	logger.Log.Infow("Starting HTTP server", "address", a.config.Addr)

	errChan := make(chan error, 1)

	go func() {
		if err := a.server.ListenAndServe(); err != nil &&
			!errors.Is(err, context.Canceled) &&
			err != http.ErrServerClosed {

			logger.Log.Errorw("HTTP server failed", "error", err)
			errChan <- err
		} else {
			logger.Log.Info("HTTP server exited gracefully")
		}
		close(errChan)
	}()

	for _, worker := range a.workers {
		go worker(ctx)
	}

	select {
	case <-ctx.Done():
		logger.Log.Info("Shutdown signal received, shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Errorw("Error during server shutdown", "error", err)
			return err
		}

		if a.db != nil {
			a.db.Close()
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

func newDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		logger.Log.Errorw("Failed to connect to database", "error", err)
		return nil, err
	}

	logger.Log.Info("Connected to database")

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Log.Errorw("Failed to set goose dialect", "error", err)
		return nil, err
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		logger.Log.Errorw("Failed to get caller info")
		return nil, err
	}
	dir := filepath.Dir(filename)

	migrationsPath := filepath.Join(dir, "..", "..", "migrations")
	migrationsPath, err = filepath.Abs(migrationsPath)
	if err != nil {
		logger.Log.Errorw("Failed to get absolute path of migrations", "error", err)
		return nil, err
	}

	if err := goose.Up(db.DB, migrationsPath); err != nil {
		logger.Log.Errorw("Failed to apply migrations", "error", err)
		return nil, err
	}

	logger.Log.Info("Database migrations applied successfully")

	return db, nil
}
