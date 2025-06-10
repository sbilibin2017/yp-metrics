package runners

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRunContext_Cancel(t *testing.T) {
	baseCtx := context.Background()
	ctx, cancel := NewRunContext(baseCtx)
	defer cancel()

	assert.NotNil(t, ctx, "context should not be nil")
	assert.NotNil(t, cancel, "cancel function should not be nil")

	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		close(done)
	}()

	// simulate a signal by calling cancel manually
	cancel()

	select {
	case <-done:
		assert.True(t, true, "context should be cancelled")
	case <-time.After(1 * time.Second):
		t.Fatal("context was not cancelled in time")
	}
}
