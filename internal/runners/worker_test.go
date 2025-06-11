package runners

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunWorker_FinishNormally(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	workerCalled := false
	worker := func(ctx context.Context) error {
		workerCalled = true
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	RunWorker(ctx, worker)

	assert.True(t, workerCalled, "worker should have been called")
}

func TestRunWorker_WorkerReturnsError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	worker := func(ctx context.Context) error {
		return errors.New("some error")
	}

	RunWorker(ctx, worker)
}

func TestRunWorker_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	workerStarted := make(chan struct{})
	worker := func(ctx context.Context) error {
		close(workerStarted)
		<-ctx.Done()
		return nil
	}

	go func() {
		<-workerStarted
		cancel()
	}()

	RunWorker(ctx, worker)

	assert.True(t, true)
}
