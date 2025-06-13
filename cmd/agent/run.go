package main

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/apps"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func run(ctx context.Context, config *configs.AgentConfig) error {
	app, err := apps.NewAgentApp(config)
	if err != nil {
		return err
	}
	return app.Start(ctx)
}
