package containers

import (
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func TestNewServerContainer(t *testing.T) {
	cfg := configs.NewServerConfig(
		configs.WithServerRunAddress(":8080"),
	)

	container, err := NewServerContainer(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, container)

	assert.NotNil(t, container.Data)
	assert.NotNil(t, container.MetricMemorySaveRepository)
	assert.NotNil(t, container.MetricMemoryGetRepository)
	assert.NotNil(t, container.MetricMemoryListRepository)

	assert.NotNil(t, container.MetricUpdateService)
	assert.NotNil(t, container.MetricGetService)
	assert.NotNil(t, container.MetricListService)

	assert.NotNil(t, container.MetricUpdatePathHandler)
	assert.NotNil(t, container.MetricGetPathHandler)
	assert.NotNil(t, container.MetricListHTMLHandler)

	assert.NotNil(t, container.MetricsRouter)
	assert.NotNil(t, container.Server)

	assert.Equal(t, ":8080", container.Server.Addr)
}
