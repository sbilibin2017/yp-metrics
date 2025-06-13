package main

import (
	"flag"
	"os"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func resetServerEnv() {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("KEY")
	os.Unsetenv("HASH_HEADER")
}

func TestServerConfigOptions(t *testing.T) {
	tests := []struct {
		name       string
		envKey     string
		envValue   string
		flagArgs   []string
		optionFunc func(fs *flag.FlagSet) configs.ServerOption
		assertFn   func(t *testing.T, cfg *configs.ServerConfig)
	}{
		{
			name:       "Addr from flag",
			envKey:     "ADDRESS",
			envValue:   "",
			flagArgs:   []string{"-a", ":9090"},
			optionFunc: withAddr,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, ":9090", cfg.Addr)
			},
		},
		{
			name:       "Addr from env",
			envKey:     "ADDRESS",
			envValue:   ":7070",
			flagArgs:   []string{},
			optionFunc: withAddr,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, ":7070", cfg.Addr)
			},
		},
		{
			name:       "StoreInterval from flag",
			envKey:     "STORE_INTERVAL",
			envValue:   "",
			flagArgs:   []string{"-i", "150"},
			optionFunc: withStoreInterval,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, 150, cfg.StoreInterval)
			},
		},
		{
			name:       "StoreInterval from env",
			envKey:     "STORE_INTERVAL",
			envValue:   "200",
			flagArgs:   []string{},
			optionFunc: withStoreInterval,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, 200, cfg.StoreInterval)
			},
		},
		{
			name:       "FileStoragePath from flag",
			envKey:     "FILE_STORAGE_PATH",
			envValue:   "",
			flagArgs:   []string{"-f", "/tmp/metrics.json"},
			optionFunc: withFileStoragePath,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "/tmp/metrics.json", cfg.FileStoragePath)
			},
		},
		{
			name:       "FileStoragePath from env",
			envKey:     "FILE_STORAGE_PATH",
			envValue:   "/env/metrics.json",
			flagArgs:   []string{},
			optionFunc: withFileStoragePath,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "/env/metrics.json", cfg.FileStoragePath)
			},
		},
		{
			name:       "Restore from flag (false)",
			envKey:     "RESTORE",
			envValue:   "",
			flagArgs:   []string{"-r=false"},
			optionFunc: withRestore,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.False(t, cfg.Restore)
			},
		},
		{
			name:       "Restore from env (true)",
			envKey:     "RESTORE",
			envValue:   "true",
			flagArgs:   []string{},
			optionFunc: withRestore,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.True(t, cfg.Restore)
			},
		},
		{
			name:       "DatabaseDSN from flag",
			envKey:     "DATABASE_DSN",
			envValue:   "",
			flagArgs:   []string{"-d", "postgres://flagdsn"},
			optionFunc: withDatabaseDSN,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "postgres://flagdsn", cfg.DatabaseDSN)
			},
		},
		{
			name:       "DatabaseDSN from env",
			envKey:     "DATABASE_DSN",
			envValue:   "postgres://envdsn",
			flagArgs:   []string{},
			optionFunc: withDatabaseDSN,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "postgres://envdsn", cfg.DatabaseDSN)
			},
		},
		{
			name:       "LogLevel from flag",
			envKey:     "LOG_LEVEL",
			envValue:   "",
			flagArgs:   []string{"-l", "debug"},
			optionFunc: withLogLevel,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "debug", cfg.LogLevel)
			},
		},
		{
			name:       "LogLevel from env",
			envKey:     "LOG_LEVEL",
			envValue:   "warn",
			flagArgs:   []string{},
			optionFunc: withLogLevel,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "warn", cfg.LogLevel)
			},
		},
		{
			name:       "HashKey from flag",
			envKey:     "KEY",
			envValue:   "",
			flagArgs:   []string{"-k", "flagkey"},
			optionFunc: withHashKey,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "flagkey", cfg.HashKey)
			},
		},
		{
			name:       "HashKey from env",
			envKey:     "KEY",
			envValue:   "envkey",
			flagArgs:   []string{},
			optionFunc: withHashKey,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "envkey", cfg.HashKey)
			},
		},
		{
			name:       "HashHeader from flag",
			envKey:     "HASH_HEADER",
			envValue:   "",
			flagArgs:   []string{"-hh", "CustomHeader"},
			optionFunc: withHashHeader,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "CustomHeader", cfg.HashHeader)
			},
		},
		{
			name:       "HashHeader from env",
			envKey:     "HASH_HEADER",
			envValue:   "EnvHeader",
			flagArgs:   []string{},
			optionFunc: withHashHeader,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "EnvHeader", cfg.HashHeader)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetServerEnv()
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
			}

			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			opt := tt.optionFunc(fs)

			err := fs.Parse(tt.flagArgs)
			assert.NoError(t, err)

			cfg := &configs.ServerConfig{}
			opt(cfg)

			tt.assertFn(t, cfg)
		})
	}
}
