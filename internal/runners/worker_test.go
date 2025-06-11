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

	workerCalled := make(chan struct{})
	worker := func(ctx context.Context) {
		time.Sleep(100 * time.Millisecond)
		close(workerCalled)
	}

	go RunWorker(ctx, worker)

	select {
	case <-workerCalled:
		assert.True(t, true, "worker should have been called")
	case <-time.After(time.Second):
		t.Fatal("worker was not called in time")
	}
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

	assert.True(t, true, "RunWorker should return after context cancellation")
}
