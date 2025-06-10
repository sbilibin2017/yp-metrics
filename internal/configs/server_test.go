package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerConfig_Default(t *testing.T) {
	cfg := NewServerConfig()

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.LogLevel)
	assert.Empty(t, cfg.RunAddress)
}

func TestWithLogLevel(t *testing.T) {
	cfg := NewServerConfig(WithServerLogLevel())

	assert.Equal(t, "info", cfg.LogLevel)
	assert.Empty(t, cfg.RunAddress)
}

func TestWithRunAddress(t *testing.T) {
	cfg := NewServerConfig(WithServerRunAddress(":9090"))

	assert.Equal(t, ":9090", cfg.RunAddress)
	assert.Empty(t, cfg.LogLevel)
}

func TestWithLogLevelAndRunAddress(t *testing.T) {
	cfg := NewServerConfig(
		WithServerLogLevel(),
		WithServerRunAddress(":8080"),
	)

	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, ":8080", cfg.RunAddress)
}
