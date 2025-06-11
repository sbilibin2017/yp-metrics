package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags_DefaultValues(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"agent"}

	cfg, err := parseFlags()
	assert.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", cfg.ServerRunAddress)
	assert.Equal(t, 2, cfg.PollInterval)
	assert.Equal(t, 10, cfg.ReportInterval)
}

func TestParseFlags_CommandLineFlags(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"agent", "-a=127.0.0.1:9000", "-p=5", "-r=15"}

	cfg, err := parseFlags()
	assert.NoError(t, err)

	assert.Equal(t, "127.0.0.1:9000", cfg.ServerRunAddress)
	assert.Equal(t, 5, cfg.PollInterval)
	assert.Equal(t, 15, cfg.ReportInterval)
}

func TestParseFlags_EnvOverrides(t *testing.T) {
	os.Clearenv()
	os.Setenv("ADDRESS", "envhost:7777") // ← исправлено
	os.Setenv("POLL_INTERVAL", "8")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Args = []string{"agent", "-a=127.0.0.1:9000", "-p=5", "-r=15"}

	cfg, err := parseFlags()
	assert.NoError(t, err)

	assert.Equal(t, "envhost:7777", cfg.ServerRunAddress)
	assert.Equal(t, 8, cfg.PollInterval)
	assert.Equal(t, 20, cfg.ReportInterval)
}

func TestParseFlags_InvalidEnvValuesFallback(t *testing.T) {
	os.Clearenv()
	os.Setenv("POLL_INTERVAL", "notanint") // некорректное значение
	os.Setenv("REPORT_INTERVAL", "20")
	os.Args = []string{"agent", "-a=127.0.0.1:9000", "-p=5", "-r=15"}

	cfg, err := parseFlags()
	assert.NoError(t, err)

	assert.Equal(t, "127.0.0.1:9000", cfg.ServerRunAddress)
	assert.Equal(t, 5, cfg.PollInterval) // fallback к флагу
	assert.Equal(t, 20, cfg.ReportInterval)
}
