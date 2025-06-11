package main

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/containers"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/runners"
)

func run(ctx context.Context, config *configs.AgentConfig) error {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	container, err := containers.NewAgentContainer(config)
	if err != nil {
		return err
	}

	ctx, stop := runners.NewRunContext(ctx)
	defer stop()

	runners.RunWorker(ctx, container.Worker)

	return nil
}
