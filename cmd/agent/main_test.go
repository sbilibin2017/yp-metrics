package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/suite"
)

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

	// Create AgentConfig explicitly instead of relying on package vars
	cfg := &configs.AgentConfig{
		Address:        s.stubServer.URL,
		PollInterval:   1,
		ReportInterval: 2,
		LogLevel:       "info",
	}

	go func() {
		err := run(ctx, cfg)
		if err != nil && err != context.Canceled {
			s.Require().NoError(err) // safe to call inside goroutine here
		}
	}()

	time.Sleep(500 * time.Millisecond) // crude wait for app start; consider sync.WaitGroup or channel

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
