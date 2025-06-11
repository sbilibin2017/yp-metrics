package main

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/sbilibin2017/yp-metrics/internal/containers"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/runners"
)

func run(ctx context.Context, config *configs.ServerConfig) error {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	ctx, stop := runners.NewRunContext(ctx)
	defer stop()

	container, err := containers.NewServerContainer(config)
	if err != nil {
		return err
	}

	err = runners.RunServer(ctx, container.Server)
	if err != nil && err != context.Canceled {
		return err
	}

	return nil
}
