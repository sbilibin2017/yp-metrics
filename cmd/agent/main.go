package main

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/facades"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/runners"
	"github.com/sbilibin2017/yp-metrics/internal/workers"
)

func main() {
	config := parseFlags()
	if err := run(context.Background(), config); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, config *configs.AgentConfig) error {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	agent, err := newAgent(config)
	if err != nil {
		return err
	}

	ctx, stop := runners.NewRunContext(ctx)
	defer stop()

	runners.RunWorker(ctx, agent)

	return nil
}

func newAgent(config *configs.AgentConfig) (func(ctx context.Context), error) {
	client := resty.New()

	metricUpdateFacade := facades.NewMetricUpdateFacade(
		client,
		config.ServerRunAddress,
		config.ServerEndpoint,
	)

	return workers.NewMetricAgentWorker(
		metricUpdateFacade,
		config.PollInterval,
		config.ReportInterval,
	), nil

}
