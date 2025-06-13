package apps

import (
	"context"
	"testing"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentApp_Success(t *testing.T) {
	cfg := &configs.AgentConfig{
		Address:        "http://localhost:8080",
		PollInterval:   1,
		ReportInterval: 1,
		LogLevel:       "info",
	}

	app, err := NewAgentApp(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, cfg, app.config)
	assert.Len(t, app.workers, 1)
}

func TestAgentApp_Start_Shutdown(t *testing.T) {
	cfg := &configs.AgentConfig{
		Address:        "http://localhost:8080",
		PollInterval:   1,
		ReportInterval: 1,
		LogLevel:       "info",
	}

	app, err := NewAgentApp(cfg)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Start the app in a goroutine
	done := make(chan error)
	go func() {
		done <- app.Start(ctx)
	}()

	// Wait a short moment then cancel to trigger shutdown
	time.Sleep(100 * time.Millisecond)
	cancel()

	err = <-done
	assert.ErrorIs(t, err, context.Canceled)
}
