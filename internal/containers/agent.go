package containers

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/facades"
	"github.com/sbilibin2017/yp-metrics/internal/workers"
)

type AgentContainer struct {
	Client             *resty.Client
	MetricUpdateFacade *facades.MetricUpdateFacade
	Worker             func(ctx context.Context) error
}

func NewAgentContainer(config *configs.AgentConfig) (*AgentContainer, error) {
	client := resty.New()

	metricUpdateFacade := facades.NewMetricUpdateFacade(
		client,
		config.ServerRunAddress,
	)

	worker := workers.NewMetricAgentWorker(
		metricUpdateFacade,
		config.PollInterval,
		config.ReportInterval,
	)

	return &AgentContainer{
		Client:             client,
		MetricUpdateFacade: metricUpdateFacade,
		Worker:             worker,
	}, nil
}
