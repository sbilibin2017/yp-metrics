package main

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/stretchr/testify/suite"
)

func resetFlags() {
	// Reset flag.CommandLine to avoid conflicts between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags()

	// Clear env vars
	os.Unsetenv("ADDRESS")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")

	// Clear os.Args to simulate no flags passed
	os.Args = []string{"cmd"}

	parseFlags()

	if address != "http://localhost:8080" {
		t.Errorf("expected default address http://localhost:8080, got %q", address)
	}
	if pollInterval != 2 {
		t.Errorf("expected default pollInterval 2, got %d", pollInterval)
	}
	if reportInterval != 10 {
		t.Errorf("expected default reportInterval 10, got %d", reportInterval)
	}
}

func TestParseFlags_WithFlags(t *testing.T) {
	resetFlags()

	os.Unsetenv("ADDRESS")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")

	// Simulate passing flags
	os.Args = []string{
		"cmd",
		"-a", "http://example.com",
		"-p", "5",
		"-r", "20",
	}

	parseFlags()

	if address != "http://example.com" {
		t.Errorf("expected address http://example.com, got %q", address)
	}
	if pollInterval != 5 {
		t.Errorf("expected pollInterval 5, got %d", pollInterval)
	}
	if reportInterval != 20 {
		t.Errorf("expected reportInterval 20, got %d", reportInterval)
	}
}

func TestParseFlags_EnvOverridesFlags(t *testing.T) {
	resetFlags()

	// Set env vars
	os.Setenv("ADDRESS", "http://env.com")
	os.Setenv("POLL_INTERVAL", "7")
	os.Setenv("REPORT_INTERVAL", "15")

	// Simulate passing different flags (which should be overridden by env)
	os.Args = []string{
		"cmd",
		"-a", "http://flag.com",
		"-p", "3",
		"-r", "6",
	}

	parseFlags()

	if address != "http://env.com" {
		t.Errorf("expected address http://env.com from env, got %q", address)
	}
	if pollInterval != 7 {
		t.Errorf("expected pollInterval 7 from env, got %d", pollInterval)
	}
	if reportInterval != 15 {
		t.Errorf("expected reportInterval 15 from env, got %d", reportInterval)
	}

	// Cleanup
	os.Unsetenv("ADDRESS")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")
}

func TestParseFlags_InvalidEnvInts(t *testing.T) {
	resetFlags()

	os.Setenv("POLL_INTERVAL", "notint")
	os.Setenv("REPORT_INTERVAL", "alsoNotInt")

	os.Args = []string{"cmd"}

	parseFlags()

	// Should fall back to defaults if env is invalid
	if pollInterval != 2 {
		t.Errorf("expected default pollInterval 2 on invalid env, got %d", pollInterval)
	}
	if reportInterval != 10 {
		t.Errorf("expected default reportInterval 10 on invalid env, got %d", reportInterval)
	}

	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")
}

type AgentSuite struct {
	suite.Suite
	cancel     context.CancelFunc
	client     *resty.Client
	stubServer *httptest.Server
}

func (s *AgentSuite) SetupSuite() {
	handler := http.NewServeMux()
	handler.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.stubServer = httptest.NewServer(handler)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	address = s.stubServer.URL
	pollInterval = 1
	reportInterval = 2
	logLevel = "info"

	go func() {
		err := run(ctx)
		if err != nil && err != context.Canceled {
			s.Require().NoError(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	s.client = resty.New().SetBaseURL(s.stubServer.URL)
}

func (s *AgentSuite) TearDownSuite() {
	s.cancel()
	s.stubServer.Close()
	time.Sleep(100 * time.Millisecond)
}

func (s *AgentSuite) TestAgentRunsAndSendsRequests() {
	resp, err := s.client.R().Post("/update/gauge/test/123.4")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
}

func TestAgentSuite(t *testing.T) {
	suite.Run(t, new(AgentSuite))
}
