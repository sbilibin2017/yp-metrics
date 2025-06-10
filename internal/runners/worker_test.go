package runners

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunWorker_FinishNormally(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	workerCalled := false
	worker := func(ctx context.Context) {
		workerCalled = true
		time.Sleep(100 * time.Millisecond)
	}

	RunWorker(ctx, worker)

	assert.True(t, workerCalled, "worker should have been called")
}

func TestRunWorker_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	workerStarted := make(chan struct{})
	worker := func(ctx context.Context) {
		close(workerStarted)
		<-ctx.Done()
	}

	go func() {
		<-workerStarted
		cancel()
	}()

	RunWorker(ctx, worker)

	assert.True(t, true)
}
