package containers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/handlers"
	"github.com/sbilibin2017/yp-metrics/internal/middlewares"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/routers"
	"github.com/sbilibin2017/yp-metrics/internal/services"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/sbilibin2017/yp-metrics/internal/validators"
)

type ServerContainer struct {
	Data map[types.MetricID]types.Metrics

	MetricMemorySaveRepository *repositories.MetricMemorySaveRepository
	MetricMemoryGetRepository  *repositories.MetricMemoryGetRepository
	MetricMemoryListRepository *repositories.MetricMemoryListRepository

	MetricUpdateService *services.MetricUpdateService
	MetricGetService    *services.MetricGetService
	MetricListService   *services.MetricListService

	MetricUpdatePathHandler http.HandlerFunc
	MetricGetPathHandler    http.HandlerFunc
	MetricListHTMLHandler   http.HandlerFunc

	MetricsRouter *chi.Mux
	Server        *http.Server
}

func NewServerContainer(config *configs.ServerConfig) (*ServerContainer, error) {
	data := make(map[types.MetricID]types.Metrics)

	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(data)
	metricMemoryGetRepository := repositories.NewMetricMemoryGetRepository(data)
	metricMemoryListRepository := repositories.NewMetricMemoryListRepository(data)

	metricUpdateService := services.NewMetricUpdateService(metricMemorySaveRepository, metricMemoryGetRepository)
	metricGetService := services.NewMetricGetService(metricMemoryGetRepository)
	metricListService := services.NewMetricListService(metricMemoryListRepository)

	metricUpdatePathHandler := handlers.MetricUpdatePathHandler(validators.ValidateMetricPath, metricUpdateService)
	metricUpdateBodyHandler := handlers.MetricUpdateBodyHandler(validators.ValidateMetricBody, metricUpdateService)
	metricGetPathHandler := handlers.MetricGetPathHandler(validators.ValidateMetricIDPath, metricGetService)
	metricGetBodyHandler := handlers.MetricGetBodyHandler(validators.ValidateMetricIDPath, metricGetService)
	metricListHTMLHandler := handlers.MetricListHTMLHandler(metricListService)

	middlewares := []func(http.Handler) http.Handler{
		middlewares.LoggingMiddleware,
	}

	metricsRouter := routers.NewMetricsRouter(
		metricUpdatePathHandler,
		metricUpdateBodyHandler,
		metricGetPathHandler,
		metricGetBodyHandler,
		metricListHTMLHandler,
		middlewares...,
	)

	srv := &http.Server{
		Addr:    config.RunAddress,
		Handler: metricsRouter,
	}

	return &ServerContainer{
		Data:                       data,
		MetricMemorySaveRepository: metricMemorySaveRepository,
		MetricMemoryGetRepository:  metricMemoryGetRepository,
		MetricMemoryListRepository: metricMemoryListRepository,
		MetricUpdateService:        metricUpdateService,
		MetricGetService:           metricGetService,
		MetricListService:          metricListService,
		MetricUpdatePathHandler:    metricUpdatePathHandler,
		MetricGetPathHandler:       metricGetPathHandler,
		MetricListHTMLHandler:      metricListHTMLHandler,
		MetricsRouter:              metricsRouter,
		Server:                     srv,
	}, nil
}
