package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func parseFlags() (*configs.AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	address := fs.String("a", "http://localhost:8080", "address of HTTP server")
	pollInterval := fs.Int("p", 2, "poll interval in seconds")
	reportInterval := fs.Int("r", 10, "report interval in seconds")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		*address = envAddr
	}
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if val, err := strconv.Atoi(envPoll); err == nil {
			*pollInterval = val
		}
	}
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		if val, err := strconv.Atoi(envReport); err == nil {
			*reportInterval = val
		}
	}

	return configs.NewAgentConfig(
		configs.WithAgentServerRunAddress(*address),
		configs.WithAgentPollInterval(*pollInterval),
		configs.WithAgentReportInterval(*reportInterval),
	), nil
}
