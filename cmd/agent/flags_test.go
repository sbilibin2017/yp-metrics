package main

import (
	"flag"
	"os"
	"testing"

	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/assert"
)

func resetAgentEnv() {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("LOG_LEVEL")
}

func TestAgentConfigOptions(t *testing.T) {
	tests := []struct {
		name       string
		envKey     string
		envValue   string
		flagArgs   []string
		optionFunc func(fs *flag.FlagSet) configs.AgentOption
		assertFn   func(t *testing.T, cfg *configs.AgentConfig)
	}{
		{
			name:       "Address from flag",
			envKey:     "ADDRESS",
			envValue:   "",
			flagArgs:   []string{"-a", "http://flagaddress:8081"},
			optionFunc: withAddress,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, "http://flagaddress:8081", cfg.Address)
			},
		},
		{
			name:       "Address from env",
			envKey:     "ADDRESS",
			envValue:   "http://envaddress:8082",
			flagArgs:   []string{},
			optionFunc: withAddress,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, "http://envaddress:8082", cfg.Address)
			},
		},
		{
			name:       "PollInterval from flag",
			envKey:     "POLL_INTERVAL",
			envValue:   "",
			flagArgs:   []string{"-p", "5"},
			optionFunc: withPollInterval,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, 5, cfg.PollInterval)
			},
		},
		{
			name:       "PollInterval from env",
			envKey:     "POLL_INTERVAL",
			envValue:   "8",
			flagArgs:   []string{},
			optionFunc: withPollInterval,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, 8, cfg.PollInterval)
			},
		},
		{
			name:       "ReportInterval from flag",
			envKey:     "REPORT_INTERVAL",
			envValue:   "",
			flagArgs:   []string{"-r", "20"},
			optionFunc: withReportInterval,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, 20, cfg.ReportInterval)
			},
		},
		{
			name:       "ReportInterval from env",
			envKey:     "REPORT_INTERVAL",
			envValue:   "30",
			flagArgs:   []string{},
			optionFunc: withReportInterval,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, 30, cfg.ReportInterval)
			},
		},
		{
			name:       "LogLevel from flag",
			envKey:     "LOG_LEVEL",
			envValue:   "",
			flagArgs:   []string{"-l", "debug"},
			optionFunc: withLogLevel,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, "debug", cfg.LogLevel)
			},
		},
		{
			name:       "LogLevel from env",
			envKey:     "LOG_LEVEL",
			envValue:   "warn",
			flagArgs:   []string{},
			optionFunc: withLogLevel,
			assertFn: func(t *testing.T, cfg *configs.AgentConfig) {
				assert.Equal(t, "warn", cfg.LogLevel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentEnv()
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
			}

			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			opt := tt.optionFunc(fs)

			err := fs.Parse(tt.flagArgs)
			assert.NoError(t, err)

			cfg := &configs.AgentConfig{}
			opt(cfg)

			tt.assertFn(t, cfg)
		})
	}
}

func TestParseFlagsEnvOverridesFlags(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		args     []string
		expected *configs.AgentConfig
	}{
		{
			name: "env overrides flags",
			env: map[string]string{
				"ADDRESS":         "http://env:8080",
				"POLL_INTERVAL":   "11",
				"REPORT_INTERVAL": "22",
				"LOG_LEVEL":       "warn",
			},
			args: []string{
				"-a", "http://flag:8080",
				"-p", "1",
				"-r", "2",
				"-l", "info",
			},
			expected: &configs.AgentConfig{
				Address:        "http://env:8080",
				PollInterval:   11,
				ReportInterval: 22,
				LogLevel:       "warn",
				HashKey:        "",
				HashHeader:     "HashSHA256",
			},
		},
		{
			name: "flags used if no env",
			env:  map[string]string{},
			args: []string{
				"-a", "http://flag:8080",
				"-p", "1",
				"-r", "2",
				"-l", "info",
			},
			expected: &configs.AgentConfig{
				Address:        "http://flag:8080",
				PollInterval:   1,
				ReportInterval: 2,
				LogLevel:       "info",
				HashKey:        "",
				HashHeader:     "HashSHA256",
			},
		},
		{
			name: "defaults used if no env or flags",
			env:  map[string]string{},
			args: []string{},
			expected: &configs.AgentConfig{
				Address:        "http://localhost:8080",
				PollInterval:   2,
				ReportInterval: 10,
				LogLevel:       "info",
				HashKey:        "",
				HashHeader:     "HashSHA256",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentEnv()
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			origArgs := os.Args
			defer func() { os.Args = origArgs }()

			os.Args = append([]string{"agent"}, tt.args...)

			cfg := parseFlags()

			assert.Equal(t, tt.expected, cfg)
		})
	}
}
