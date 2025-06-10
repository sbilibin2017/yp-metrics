package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAgentConfig_Default(t *testing.T) {
	cfg := NewAgentConfig()

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.LogLevel)
	assert.Empty(t, cfg.ServerRunAddress)
	assert.Empty(t, cfg.PollInterval)
	assert.Empty(t, cfg.ReportInterval)
}

func TestWithAgentLogLevel(t *testing.T) {
	cfg := NewAgentConfig(WithAgentLogLevel())

	assert.Equal(t, "info", cfg.LogLevel)
	assert.Empty(t, cfg.ServerRunAddress)
	assert.Empty(t, cfg.PollInterval)
	assert.Empty(t, cfg.ReportInterval)
}

func TestWithAgentServerRunAddress(t *testing.T) {
	cfg := NewAgentConfig(WithAgentServerRunAddress(":9090"))

	assert.Equal(t, ":9090", cfg.ServerRunAddress)
	assert.Empty(t, cfg.LogLevel)
	assert.Empty(t, cfg.PollInterval)
	assert.Empty(t, cfg.ReportInterval)
}

func TestWithAgentPollInterval(t *testing.T) {
	cfg := NewAgentConfig(WithAgentPollInterval(10))

	assert.Equal(t, 10, cfg.PollInterval)
	assert.Empty(t, cfg.LogLevel)
	assert.Empty(t, cfg.ServerRunAddress)
	assert.Empty(t, cfg.ReportInterval)
}

func TestWithAgentReportInterval(t *testing.T) {
	cfg := NewAgentConfig(WithAgentReportInterval(20))

	assert.Equal(t, 20, cfg.ReportInterval)
	assert.Empty(t, cfg.LogLevel)
	assert.Empty(t, cfg.ServerRunAddress)
	assert.Empty(t, cfg.PollInterval)
}

func TestWithMultipleAgentOptions(t *testing.T) {
	cfg := NewAgentConfig(
		WithAgentLogLevel(),
		WithAgentServerRunAddress(":8080"),
		WithAgentPollInterval(15),
		WithAgentReportInterval(30),
	)

	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, ":8080", cfg.ServerRunAddress)
	assert.Equal(t, 15, cfg.PollInterval)
	assert.Equal(t, 30, cfg.ReportInterval)
}

func TestWithAgentServerEndpoint(t *testing.T) {
	cfg := NewAgentConfig(WithAgentServerEndpoint())

	assert.Equal(t, "/update", cfg.ServerEndpoint)
	assert.Empty(t, cfg.LogLevel)
	assert.Empty(t, cfg.ServerRunAddress)
	assert.Empty(t, cfg.PollInterval)
	assert.Empty(t, cfg.ReportInterval)
}
