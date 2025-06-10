package runners

import (
	"context"
	"errors"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRunServer_GracefulShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockServer(ctrl)

	// Expect ListenAndServe to block until shutdown (simulated by returning http.ErrServerClosed)
	mockServer.EXPECT().ListenAndServe().DoAndReturn(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	mockServer.EXPECT().Shutdown(gomock.Any()).Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := RunServer(ctx, mockServer)
	assert.Equal(t, context.Canceled, err)
}

func TestRunServer_ListenAndServeFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockServer(ctrl)

	expectedErr := errors.New("server failed to start")

	mockServer.EXPECT().ListenAndServe().Return(expectedErr)

	ctx := context.Background()
	err := RunServer(ctx, mockServer)

	assert.Equal(t, expectedErr, err)
}

func TestRunServer_ShutdownFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockServer(ctrl)

	mockServer.EXPECT().ListenAndServe().DoAndReturn(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	shutdownErr := errors.New("shutdown failed")
	mockServer.EXPECT().Shutdown(gomock.Any()).Return(shutdownErr)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := RunServer(ctx, mockServer)
	assert.Equal(t, shutdownErr, err)
}
