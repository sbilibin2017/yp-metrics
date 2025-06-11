package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/suite"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags()
	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")

	// Simulate no command-line flags
	os.Args = []string{"cmd"}

	parseFlags()

	if addr != ":8080" {
		t.Errorf("expected default addr ':8080', got %q", addr)
	}
	if storeInterval != 300 {
		t.Errorf("expected default storeInterval 300, got %d", storeInterval)
	}
	if fileStoragePath != "./data/metrics.json" {
		t.Errorf("expected default fileStoragePath './data/metrics.json', got %q", fileStoragePath)
	}
	if restore != true {
		t.Errorf("expected default restore true, got %v", restore)
	}
}

func TestParseFlags_WithFlags(t *testing.T) {
	resetFlags()
	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")

	os.Args = []string{
		"cmd",
		"-a", ":9090",
		"-i", "100",
		"-f", "/tmp/metrics.json",
		"-r=false",
	}

	parseFlags()

	if addr != ":9090" {
		t.Errorf("expected addr ':9090', got %q", addr)
	}
	if storeInterval != 100 {
		t.Errorf("expected storeInterval 100, got %d", storeInterval)
	}
	if fileStoragePath != "/tmp/metrics.json" {
		t.Errorf("expected fileStoragePath '/tmp/metrics.json', got %q", fileStoragePath)
	}
	if restore != false {
		t.Errorf("expected restore false, got %v", restore)
	}
}

func TestParseFlags_EnvOverridesFlags(t *testing.T) {
	resetFlags()

	os.Setenv("ADDRESS", ":7070")
	os.Setenv("STORE_INTERVAL", "150")
	os.Setenv("FILE_STORAGE_PATH", "/env/metrics.json")
	os.Setenv("RESTORE", "false")

	os.Args = []string{
		"cmd",
		"-a", ":9090",
		"-i", "100",
		"-f", "/tmp/metrics.json",
		"-r=true",
	}

	parseFlags()

	if addr != ":7070" {
		t.Errorf("expected addr ':7070' from env, got %q", addr)
	}
	if storeInterval != 150 {
		t.Errorf("expected storeInterval 150 from env, got %d", storeInterval)
	}
	if fileStoragePath != "/env/metrics.json" {
		t.Errorf("expected fileStoragePath '/env/metrics.json' from env, got %q", fileStoragePath)
	}
	if restore != false {
		t.Errorf("expected restore false from env, got %v", restore)
	}

	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")
}

func TestParseFlags_InvalidEnvValues(t *testing.T) {
	resetFlags()

	// Set invalid environment variables to check fallback behavior
	os.Setenv("STORE_INTERVAL", "invalidint")
	os.Setenv("RESTORE", "notabool")

	os.Args = []string{"cmd"}

	parseFlags()

	// Expect defaults since invalid env values should not override
	if storeInterval != 300 {
		t.Errorf("expected default storeInterval 300 with invalid env, got %d", storeInterval)
	}
	if restore != true {
		t.Errorf("expected default restore true with invalid env, got %v", restore)
	}

	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("RESTORE")
}

type MainSuite struct {
	suite.Suite
	cancel context.CancelFunc
	client *resty.Client
}

func (s *MainSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// Set globals expected by run()
	addr = ":8080"
	storeInterval = 1 // or your preferred test value
	fileStoragePath = "./testdata/metrics.json"
	restore = false
	logLevel = "info"

	go func() {
		err := run(ctx)
		if err != nil && err != context.Canceled {
			s.Require().NoError(err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	s.client = resty.New().SetBaseURL("http://localhost:8080")
}

func (s *MainSuite) TearDownSuite() {
	s.cancel()
	time.Sleep(50 * time.Millisecond)

	err := os.RemoveAll("./testdata")
	s.Require().NoError(err)
}

func (s *MainSuite) TestUpdateMetric() {
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		metricValue    string
		expectedStatus int
	}{
		{
			name:           "Valid gauge metric",
			metricType:     "gauge",
			metricName:     "temperature",
			metricValue:    "42",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid counter metric",
			metricType:     "counter",
			metricName:     "requests",
			metricValue:    "100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid metric type",
			metricType:     "invalid",
			metricName:     "some",
			metricValue:    "10",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid value",
			metricType:     "gauge",
			metricName:     "pressure",
			metricValue:    "notanumber",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Value missing",
			metricType:     "gauge",
			metricName:     "temperature",
			metricValue:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := "/update/" + tt.metricType + "/" + tt.metricName
			if tt.metricValue != "" {
				url += "/" + tt.metricValue
			}
			resp, err := s.client.R().Post(url)

			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode())
		})
	}
}

func (s *MainSuite) TestGetMetricValue() {
	// Создадим метрики, чтобы потом получить их значения
	_, err := s.client.R().Post("/update/gauge/temperature/42")
	s.Require().NoError(err)

	_, err = s.client.R().Post("/update/counter/requests/100")
	s.Require().NoError(err)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Get gauge metric",
			url:            "/value/gauge/temperature",
			expectedStatus: http.StatusOK,
			expectedBody:   "42",
		},
		{
			name:           "Get counter metric",
			url:            "/value/counter/requests",
			expectedStatus: http.StatusOK,
			expectedBody:   "100",
		},
		{
			name:           "Metric not found",
			url:            "/value/gauge/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "Invalid metric type",
			url:            "/value/invalid/name",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp, err := s.client.R().Get(tt.url)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode())
			if tt.expectedBody != "" {
				s.Equal(tt.expectedBody, resp.String())
			}
		})
	}
}

func (s *MainSuite) TestGetMetricsListHTML() {
	// Добавим метрики для отображения
	_, err := s.client.R().Post("/update/gauge/temperature/42")
	s.Require().NoError(err)
	_, err = s.client.R().Post("/update/counter/requests/100")
	s.Require().NoError(err)

	resp, err := s.client.R().Get("/")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())

}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}
