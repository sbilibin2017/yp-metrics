package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
	cancel context.CancelFunc
	client *resty.Client
}

func (s *MainSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	errCh := make(chan error, 1)

	go func() {
		err := run(ctx)
		if err != nil && err != context.Canceled {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// Wait a short time for server startup
	time.Sleep(200 * time.Millisecond)

	// Check if the goroutine reported any startup error
	select {
	case err := <-errCh:
		s.Require().NoError(err)
	default:
		// no error reported yet (server likely running)
	}

	s.client = resty.New().SetBaseURL("http://localhost:8080")
}

func (s *MainSuite) TearDownSuite() {
	s.cancel()
	time.Sleep(50 * time.Millisecond)
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
			expectedStatus: http.StatusBadRequest, // Assuming your router rejects unknown types
		},
		{
			name:           "Invalid value",
			metricType:     "gauge",
			metricName:     "pressure",
			metricValue:    "notanumber",
			expectedStatus: http.StatusBadRequest, // Assuming handler rejects bad values
		},
		{
			name:           "Value missing",
			metricType:     "gauge",
			metricName:     "temperature",
			metricValue:    "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := "/update/" + tt.metricType + "/" + tt.metricName + "/" + tt.metricValue
			resp, err := s.client.R().Post(url)

			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode())
		})
	}
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}
