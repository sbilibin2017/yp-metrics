package apps

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/facades"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/workers"
)

type AgentApp struct {
	config  *configs.AgentConfig
	workers []func(ctx context.Context)
}

func NewAgentApp(cfg *configs.AgentConfig) (*AgentApp, error) {
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return nil, err
	}

	client := resty.New()
	metricFacade := facades.NewMetricUpdateFacade(client, cfg.Address)

	workersList := []func(ctx context.Context){
		func(ctx context.Context) {
			workers.StartMetricAgentWorker(ctx, metricFacade, cfg.PollInterval, cfg.ReportInterval)
		},
	}

	return &AgentApp{
		config:  cfg,
		workers: workersList,
	}, nil
}

func (a *AgentApp) Start(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	for _, worker := range a.workers {
		go worker(ctx)
	}

	<-ctx.Done()

	logger.Log.Info("Shutdown signal received, stopping agent")

	return ctx.Err()
}
