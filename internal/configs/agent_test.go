package configs_test

import (
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentConfig(t *testing.T) {
	tests := []struct {
		name   string
		opts   []configs.AgentOption
		verify func(cfg *configs.AgentConfig)
	}{
		{
			name: "no options sets defaults",
			opts: nil,
			verify: func(cfg *configs.AgentConfig) {
				assert.Equal(t, "", cfg.Address)
				assert.Equal(t, 0, cfg.PollInterval)
				assert.Equal(t, 0, cfg.ReportInterval)
				assert.Equal(t, "", cfg.LogLevel)
				assert.Equal(t, "", cfg.HashKey)
			},
		},
		{
			name: "set Address option",
			opts: []configs.AgentOption{
				func(cfg *configs.AgentConfig) {
					cfg.Address = "localhost:8080"
				},
			},
			verify: func(cfg *configs.AgentConfig) {
				assert.Equal(t, "localhost:8080", cfg.Address)
			},
		},
		{
			name: "set multiple options",
			opts: []configs.AgentOption{
				func(cfg *configs.AgentConfig) {
					cfg.Address = "127.0.0.1:9000"
					cfg.PollInterval = 15
				},
				func(cfg *configs.AgentConfig) {
					cfg.ReportInterval = 30
					cfg.LogLevel = "info"
					cfg.HashKey = "secret-key"
				},
			},
			verify: func(cfg *configs.AgentConfig) {
				assert.Equal(t, "127.0.0.1:9000", cfg.Address)
				assert.Equal(t, 15, cfg.PollInterval)
				assert.Equal(t, 30, cfg.ReportInterval)
				assert.Equal(t, "info", cfg.LogLevel)
				assert.Equal(t, "secret-key", cfg.HashKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := configs.NewAgentConfig(tt.opts...)
			tt.verify(cfg)
		})
	}
}
