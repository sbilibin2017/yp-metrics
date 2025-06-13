package configs_test

import (
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func TestNewServerConfig(t *testing.T) {
	tests := []struct {
		name   string
		opts   []configs.ServerOption
		verify func(cfg *configs.ServerConfig)
	}{
		{
			name: "no options sets defaults",
			opts: nil,
			verify: func(cfg *configs.ServerConfig) {
				assert.Equal(t, "", cfg.Addr)
				assert.Equal(t, 0, cfg.StoreInterval)
				assert.Equal(t, "", cfg.FileStoragePath)
				assert.False(t, cfg.Restore)
				assert.Equal(t, "", cfg.DatabaseDSN)
				assert.Equal(t, "", cfg.LogLevel)
			},
		},
		{
			name: "set Addr option",
			opts: []configs.ServerOption{
				func(cfg *configs.ServerConfig) {
					cfg.Addr = ":8080"
				},
			},
			verify: func(cfg *configs.ServerConfig) {
				assert.Equal(t, ":8080", cfg.Addr)
			},
		},
		{
			name: "set multiple options",
			opts: []configs.ServerOption{
				func(cfg *configs.ServerConfig) {
					cfg.Addr = "localhost:9000"
					cfg.StoreInterval = 100
				},
				func(cfg *configs.ServerConfig) {
					cfg.Restore = true
					cfg.LogLevel = "debug"
				},
			},
			verify: func(cfg *configs.ServerConfig) {
				assert.Equal(t, "localhost:9000", cfg.Addr)
				assert.Equal(t, 100, cfg.StoreInterval)
				assert.True(t, cfg.Restore)
				assert.Equal(t, "debug", cfg.LogLevel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := configs.NewServerConfig(tt.opts...)
			tt.verify(cfg)
		})
	}
}
