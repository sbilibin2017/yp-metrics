package apps_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/yp-metrics/internal/apps"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
)

func TestNewServerApp_Success(t *testing.T) {
	cfg := &configs.ServerConfig{
		Addr:            ":0", // OS chooses port
		FileStoragePath: "",
		DatabaseDSN:     "",
		LogLevel:        "info",
		StoreInterval:   1,
		Restore:         false,
	}

	app, err := apps.NewServerApp(cfg)
	require.NoError(t, err)
	require.NotNil(t, app)
}

func TestNewServerApp_BadDB(t *testing.T) {
	cfg := &configs.ServerConfig{
		Addr:        ":0",
		DatabaseDSN: "invalid-dsn",
		LogLevel:    "info",
	}

	app, err := apps.NewServerApp(cfg)
	assert.Error(t, err)
	assert.Nil(t, app)
}

func TestStart_GracefulShutdown(t *testing.T) {
	cfg := &configs.ServerConfig{
		Addr:            ":0",
		FileStoragePath: "",
		LogLevel:        "debug",
	}

	app, err := apps.NewServerApp(cfg)
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- app.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	cancel() // trigger shutdown

	err = <-done
	assert.ErrorIs(t, err, context.Canceled)
}

func TestStart_ServerError(t *testing.T) {
	cfg := &configs.ServerConfig{
		Addr:     ":99999", // invalid port to cause ListenAndServe error
		LogLevel: "info",
	}

	app, err := apps.NewServerApp(cfg)
	require.NoError(t, err)
	require.NotNil(t, app)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- app.Start(ctx)
	}()

	select {
	case err := <-errCh:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for server error")
	}
}
