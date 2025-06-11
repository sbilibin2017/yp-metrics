package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags_DefaultValues(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"server"}

	cfg, err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, ":8080", cfg.RunAddress)
}

func TestParseFlags_CommandLineFlags(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"server", "-a=127.0.0.1:9000"}

	cfg, err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9000", cfg.RunAddress)
}

func TestParseFlags_EnvOverrides(t *testing.T) {
	os.Clearenv()
	os.Setenv("ADDRESS", "envhost:7777")
	os.Args = []string{"server", "-a=127.0.0.1:9000"}

	cfg, err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, "envhost:7777", cfg.RunAddress)
}
