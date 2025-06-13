package main

import (
	"context"
	"testing"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func Test_run_basic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := &configs.ServerConfig{
		Addr:            ":0",
		StoreInterval:   1,
		FileStoragePath: "",
		Restore:         false,
		DatabaseDSN:     "",
		LogLevel:        "info",
	}

	errCh := make(chan error)
	go func() {
		errCh <- run(ctx, cfg)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	err := <-errCh
	if err != nil && err != context.Canceled {
		t.Errorf("run returned error: %v", err)
	}
}
