package main

import (
	"flag"
	"os"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func parseFlags() *configs.ServerConfig {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	var addr string
	fs.StringVar(&addr, "a", ":8080", "address and port to run server")

	_ = fs.Parse(os.Args[1:])

	if envRunAddr := os.Getenv("RUN_ADDR"); envRunAddr != "" {
		addr = envRunAddr
	}

	return configs.NewServerConfig(
		configs.WithServerRunAddress(addr),
		configs.WithServerLogLevel(),
	)
}
