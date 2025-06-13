package main

import (
	"context"

	"github.com/sbilibin2017/yp-metrics/internal/apps"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func run(ctx context.Context, config *configs.ServerConfig) error {
	app, err := apps.NewServerApp(config)
	if err != nil {
		return err
	}
	return app.Start(ctx)
}
