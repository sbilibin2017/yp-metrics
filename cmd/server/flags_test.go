package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags_DefaultValues(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"server"}

	cfg := parseFlags()

	assert.Equal(t, ":8080", cfg.RunAddress)

}

func TestParseFlags_CommandLineFlags(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"server", "-a=127.0.0.1:9000"}

	cfg := parseFlags()

	assert.Equal(t, "127.0.0.1:9000", cfg.RunAddress)

}

func TestParseFlags_EnvOverrides(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"server", "-a=127.0.0.1:9000"}

	os.Setenv("RUN_ADDR", "envhost:7777")

	cfg := parseFlags()

	assert.Equal(t, "envhost:7777", cfg.RunAddress)

}

func TestParseFlags_InvalidEnvValuesFallback(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"server", "-a=127.0.0.1:9000"}

	os.Setenv("POLL_INTERVAL", "notanint") // некорректное значение
	os.Setenv("REPORT_INTERVAL", "20")

	cfg := parseFlags()

	assert.Equal(t, "127.0.0.1:9000", cfg.RunAddress)

}
