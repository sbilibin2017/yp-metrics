package runners

import (
	"context"
)

func RunWorker(ctx context.Context, worker func(ctx context.Context)) {
	go func() {
		worker(ctx)
	}()
	<-ctx.Done()
}
