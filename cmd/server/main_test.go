package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

	// Set global DSN for your app if needed
	databaseDSN = dsn

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// Set your app globals here
	addr = ":8080"
	storeInterval = 1
	fileStoragePath = "./testdata/metrics.json"
	restore = false
	logLevel = "info"

	go func() {
		err := run(ctx)
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
	// First update the metric
	_, err := s.client.R().Post("/update/gauge/testmetric/123.45")
	s.Require().NoError(err)

	// Then get it
	resp, err := s.client.R().Get("/value/gauge/testmetric")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
	s.Contains(resp.String(), "123.45")
}

func (s *MainSuite) TestGetMetricWithBody() {
	// First update the metric
	_, err := s.client.R().Post("/update/gauge/memusage/64.0")
	s.Require().NoError(err)

	// Request metric using JSON
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
	// Добавим метрики для отображения
	_, err := s.client.R().Post("/update/gauge/temperature/42")
	s.Require().NoError(err)
	_, err = s.client.R().Post("/update/counter/requests/100")
	s.Require().NoError(err)

	resp, err := s.client.R().Get("/")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())

}

func TestPingHandler_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "pgx")

	mock.ExpectPing()

	handler := pingHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPingHandler_DBError(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "pgx")

	mock.ExpectPing().WillReturnError(context.DeadlineExceeded)

	handler := pingHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainSuite))
}
