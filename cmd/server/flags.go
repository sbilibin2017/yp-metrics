package main

import (
	"flag"
	"os"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func parseFlags() (*configs.ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	addr := fs.String("a", ":8080", "address and port to run server")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		*addr = envAddr
	}

	return configs.NewServerConfig(
		configs.WithServerRunAddress(*addr),
		configs.WithServerLogLevel(),
	), nil
}
