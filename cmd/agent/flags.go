package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func parseFlags() *configs.AgentConfig {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var (
		serverRunAddress string
		pollInterval     int
		reportInterval   int
	)

	fs.StringVar(&serverRunAddress, "a", "localhost:8080", "address and port to run server")
	fs.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	fs.IntVar(&reportInterval, "r", 10, "report interval in seconds")

	_ = fs.Parse(os.Args[1:])

	if envRunAddr := os.Getenv("RUN_ADDR"); envRunAddr != "" {
		serverRunAddress = envRunAddr
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if val, err := strconv.Atoi(envPollInterval); err == nil {
			pollInterval = val
		}
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if val, err := strconv.Atoi(envReportInterval); err == nil {
			reportInterval = val
		}
	}

	cfg := configs.NewAgentConfig(
		configs.WithAgentServerRunAddress(serverRunAddress),
		configs.WithAgentPollInterval(pollInterval),
		configs.WithAgentReportInterval(reportInterval),
	)

	return cfg
}
