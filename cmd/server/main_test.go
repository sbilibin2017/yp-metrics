package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/yp-metrics/internal/configs"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MainSuite struct {
	suite.Suite
	cancel      context.CancelFunc
	client      *resty.Client
	pgContainer testcontainers.Container
	db          *sqlx.DB
}

func (s *MainSuite) SetupSuite() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.pgContainer = pgC

	host, err := pgC.Host(ctx)
	s.Require().NoError(err)

	port, err := pgC.MappedPort(ctx, "5432")
	s.Require().NoError(err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	var db *sqlx.DB
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("pgx", dsn)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	s.Require().NoError(err)
	s.db = db

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	config := &configs.ServerConfig{
		Addr:            ":8080",
		StoreInterval:   1,
		FileStoragePath: "./testdata/metrics.json",
		Restore:         false,
		DatabaseDSN:     dsn,
		LogLevel:        "info",
	}

	go func() {
		err := run(ctx, config)
		if err != nil && err != context.Canceled {
			s.Require().NoError(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	s.client = resty.New().SetBaseURL("http://localhost:8080")
}

func (s *MainSuite) TearDownSuite() {
	s.cancel()
	time.Sleep(50 * time.Millisecond)

	if s.db != nil {
		s.db.Close()
	}

	if s.pgContainer != nil {
		s.pgContainer.Terminate(context.Background())
	}

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
		})
	}
}

func (s *MainSuite) TestUpdateMetricWithBody() {
	body := map[string]interface{}{
		"id":    "cpu_load",
		"type":  "gauge",
		"value": 75.5,
	}

	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/update/")

	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
}

func (s *MainSuite) TestGetMetricPath() {
	_, err := s.client.R().Post("/update/gauge/testmetric/123.45")
	s.Require().NoError(err)

	resp, err := s.client.R().Get("/value/gauge/testmetric")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
	s.Contains(resp.String(), "123.45")
}

func (s *MainSuite) TestGetMetricWithBody() {
	_, err := s.client.R().Post("/update/gauge/memusage/64.0")
	s.Require().NoError(err)

	body := map[string]interface{}{
		"id":   "memusage",
		"type": "gauge",
	}
	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/value/")

	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())

	var metric struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Value *float64 `json:"value"`
	}
	err = json.Unmarshal(resp.Body(), &metric)
	s.Require().NoError(err)
	s.Equal("memusage", metric.ID)
	s.Equal("gauge", metric.MType)
	s.NotNil(metric.Value)
}

func (s *MainSuite) TestGetMetricsListHTML() {
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
