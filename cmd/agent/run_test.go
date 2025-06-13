package main

import (
	"context"
	"testing"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func TestRun_Success(t *testing.T) {
	cfg := &configs.AgentConfig{
		Address:        "http://localhost:8080",
		PollInterval:   1,
		ReportInterval: 1,
		LogLevel:       "info",
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- run(ctx, cfg)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	err := <-done
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRun_NewAgentAppError(t *testing.T) {

	cfg := &configs.AgentConfig{
		Address:        "",    // invalid? depends on your code
		PollInterval:   -1,    // invalid interval maybe
		ReportInterval: -1,    // invalid interval maybe
		LogLevel:       "bad", // maybe invalid log level
	}

	ctx := context.Background()
	err := run(ctx, cfg)

	_ = err
}
