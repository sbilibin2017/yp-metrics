package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-resty/resty/v2"

	"github.com/sbilibin2017/yp-metrics/internal/facades"
	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"github.com/sbilibin2017/yp-metrics/internal/workers"
)

func main() {
	parseFlags()
	err := run(context.Background())
	if err != nil {
		panic(err)
	}
}

var (
	address        string
	pollInterval   int
	reportInterval int
	logLevel       string = "info"
)

func parseFlags() {
	flag.StringVar(&address, "a", "http://localhost:8080", "address of HTTP server")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = envAddr
	}

	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if val, err := strconv.Atoi(envPoll); err == nil {
			pollInterval = val
		}
	}

	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		if val, err := strconv.Atoi(envReport); err == nil {
			reportInterval = val
		}
	}

}

func run(ctx context.Context) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	client := resty.New()

	metricUpdateFacade := facades.NewMetricUpdateFacade(
		client,
		address,
	)

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	go workers.StartMetricAgentWorker(
		ctx,
		metricUpdateFacade,
		pollInterval,
		reportInterval,
	)

	<-ctx.Done()
	logger.Log.Info("Shutdown signal received, stopping agent worker")

	return ctx.Err()
}
