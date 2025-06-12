package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pressly/goose"
	"github.com/sbilibin2017/yp-metrics/internal/contexts"
	"github.com/sbilibin2017/yp-metrics/internal/handlers"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/middlewares"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/services"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
	"github.com/sbilibin2017/yp-metrics/internal/workers"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
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
	databaseDSN     string
)

var (
	db       *sqlx.DB
	err      error
	logLevel string = "info"
)

func parseFlags() {
	flag.StringVar(&addr, "a", ":8080", "address and port to run server")
	flag.IntVar(&storeInterval, "i", 300, "store interval in seconds (0 = synchronous write)")
	flag.StringVar(&fileStoragePath, "f", "./data/metrics.json", "file storage path")
	flag.BoolVar(&restore, "r", true, "restore metrics from file at startup")
	flag.StringVar(&databaseDSN, "d", "", "PostgreSQL DSN")

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
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		databaseDSN = envDSN
	}
}

func run(ctx context.Context) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	logger.Log.Infow("Starting application", "addr", addr, "storeInterval", storeInterval,
		"fileStoragePath", fileStoragePath, "restore", restore, "databaseDSN_set", databaseDSN != "")

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	data := make(map[types.MetricID]types.Metrics)

	if databaseDSN != "" {
		db, err = newDB(ctx, databaseDSN)
		if err != nil {
			return err
		}
	}

	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(data)
	metricFileSaveRepository := repositories.NewMetricFileSaveRepository(fileStoragePath)
	metricDBSaveRepository := repositories.NewMetricDBSaveRepository(db, contexts.GetTxFromContext)

	metricMemoryGetRepository := repositories.NewMetricMemoryGetRepository(data)
	metricFileGetRepository := repositories.NewMetricFileGetRepository(fileStoragePath)
	metricDBGetRepository := repositories.NewMetricDBGetRepository(db, contexts.GetTxFromContext)

	metricMemoryListRepository := repositories.NewMetricMemoryListRepository(data)
	metricFileListRepository := repositories.NewMetricFileListRepository(fileStoragePath)
	metricDBListRepository := repositories.NewMetricDBListRepository(db, contexts.GetTxFromContext)

	metricSaverContext := repositories.NewMetricSaverContext()
	metricGetterContext := repositories.NewMetricGetterContext()
	metricListerContext := repositories.NewMetricListerContext()

	if databaseDSN != "" {
		metricSaverContext.SetContext(metricDBSaveRepository)
		metricGetterContext.SetContext(metricDBGetRepository)
		metricListerContext.SetContext(metricDBListRepository)
		logger.Log.Info("Using database repositories for saver, getter, and lister")
	} else if fileStoragePath != "" {
		metricSaverContext.SetContext(metricFileSaveRepository)
		metricGetterContext.SetContext(metricFileGetRepository)
		metricListerContext.SetContext(metricFileListRepository)
		logger.Log.Infow("Using file repositories for saver, getter, and lister", "fileStoragePath", fileStoragePath)
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
	metricGetPathHandler := handlers.MetricGetPathHandler(validators.ValidateMetricIDPath, metricGetService)
	metricGetBodyHandler := handlers.MetricGetBodyHandler(validators.ValidateMetricIDPath, metricGetService)
	metricListHTMLHandler := handlers.MetricListHTMLHandler(metricListService)

	metricRouter := newMetricRouter(
		metricUpdatePathHandler,
		metricUpdateBodyHandler,
		metricGetPathHandler,
		metricGetBodyHandler,
		metricListHTMLHandler,
	)

	router := chi.NewRouter()
	router.Mount("/", metricRouter)
	router.Get("/ping", pingHandler(db))

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	logger.Log.Infow("Starting HTTP server", "address", addr)
	errChan := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, context.Canceled) && err != http.ErrServerClosed {
			logger.Log.Errorw("HTTP server failed", "error", err)
			errChan <- err
		} else {
			logger.Log.Info("HTTP server exited gracefully")
		}
		close(errChan)
	}()

	go workers.StartMetricServerWorker(
		ctx,
		metricMemorySaveRepository,
		metricFileSaveRepository,
		metricMemoryListRepository,
		metricFileListRepository,
		storeInterval,
		restore,
	)

	select {
	case <-ctx.Done():
		logger.Log.Info("Shutdown signal received, shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Errorw("Error during server shutdown", "error", err)
			return err
		}

		if db != nil {
			db.Close()
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

func newMetricRouter(
	metricUpdatePathHandler http.HandlerFunc,
	metricUpdateBodyHandler http.HandlerFunc,
	metricGetPathHandler http.HandlerFunc,
	metricGetBodyHandler http.HandlerFunc,
	metricListHTMLHandler http.HandlerFunc,
) *chi.Mux {
	middlewares := []func(http.Handler) http.Handler{
		middlewares.LoggingMiddleware,
		middlewares.GzipMiddleware,
		middlewares.TxMiddleware(db, contexts.SetTxToContext),
	}

	router := chi.NewRouter()
	router.Use(middlewares...)

	router.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	router.Post("/update/{type}/{name}", metricUpdatePathHandler)
	router.Post("/update/", metricUpdateBodyHandler)

	router.Get("/value/{type}/{name}", metricGetPathHandler)
	router.Get("/value/{type}", metricGetPathHandler)
	router.Post("/value/", metricGetBodyHandler)

	router.Get("/", metricListHTMLHandler)

	return router
}

func pingHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func newDB(ctx context.Context, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
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
