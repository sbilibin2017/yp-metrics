package main

import (
	"flag"
	"os"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func resetEnv() {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("LOG_LEVEL")
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
			flagArgs:   []string{"-a", "localhost:9000"},
			optionFunc: withAddr,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "localhost:9000", cfg.Addr)
			},
		},
		{
			name:       "Addr from env",
			envKey:     "ADDRESS",
			envValue:   "127.0.0.1:9999",
			flagArgs:   []string{},
			optionFunc: withAddr,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "127.0.0.1:9999", cfg.Addr)
			},
		},
		{
			name:       "StoreInterval from flag",
			envKey:     "STORE_INTERVAL",
			envValue:   "",
			flagArgs:   []string{"-i", "42"},
			optionFunc: withStoreInterval,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, 42, cfg.StoreInterval)
			},
		},
		{
			name:       "StoreInterval from env",
			envKey:     "STORE_INTERVAL",
			envValue:   "77",
			flagArgs:   []string{},
			optionFunc: withStoreInterval,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, 77, cfg.StoreInterval)
			},
		},
		{
			name:       "FileStoragePath from flag",
			envKey:     "FILE_STORAGE_PATH",
			envValue:   "",
			flagArgs:   []string{"-f", "/tmp/file.json"},
			optionFunc: withFileStoragePath,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "/tmp/file.json", cfg.FileStoragePath)
			},
		},
		{
			name:       "FileStoragePath from env",
			envKey:     "FILE_STORAGE_PATH",
			envValue:   "/env/path.json",
			flagArgs:   []string{},
			optionFunc: withFileStoragePath,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "/env/path.json", cfg.FileStoragePath)
			},
		},
		{
			name:       "Restore from flag",
			envKey:     "RESTORE",
			envValue:   "",
			flagArgs:   []string{"-r=false"},
			optionFunc: withRestore,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.False(t, cfg.Restore)
			},
		},
		{
			name:       "Restore from env",
			envKey:     "RESTORE",
			envValue:   "false",
			flagArgs:   []string{},
			optionFunc: withRestore,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.False(t, cfg.Restore)
			},
		},
		{
			name:       "DatabaseDSN from flag",
			envKey:     "DATABASE_DSN",
			envValue:   "",
			flagArgs:   []string{"-d", "user:pass@tcp(host:5432)/db"},
			optionFunc: withDatabaseDSN,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "user:pass@tcp(host:5432)/db", cfg.DatabaseDSN)
			},
		},
		{
			name:       "DatabaseDSN from env",
			envKey:     "DATABASE_DSN",
			envValue:   "postgres://env@localhost/db",
			flagArgs:   []string{},
			optionFunc: withDatabaseDSN,
			assertFn: func(t *testing.T, cfg *configs.ServerConfig) {
				assert.Equal(t, "postgres://env@localhost/db", cfg.DatabaseDSN)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetEnv()
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
