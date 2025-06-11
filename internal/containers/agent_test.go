package containers

import (
	"context"
	"testing"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentContainer(t *testing.T) {
	cfg := configs.NewAgentConfig(
		configs.WithAgentServerRunAddress("http://localhost:8080"),
		configs.WithAgentPollInterval(2),
		configs.WithAgentReportInterval(10),
	)

	container, err := NewAgentContainer(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	// Проверяем зависимости
	assert.NotNil(t, container.Client)
	assert.NotNil(t, container.MetricUpdateFacade)
	assert.NotNil(t, container.Worker)

	// Проверим, что worker запускается без паники
	done := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Worker panicked: %v", r)
			}
			close(done)
		}()

		container.Worker(ctx)
	}()
	<-done
}
