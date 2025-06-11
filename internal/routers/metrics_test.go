package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func dummyHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func TestNewMetricsRouter(t *testing.T) {
	updateHandler := dummyHandler(http.StatusOK, "update OK")
	getPathHandler := dummyHandler(http.StatusOK, "get path OK")
	getBodyHandler := dummyHandler(http.StatusOK, "get body OK")
	listHandler := dummyHandler(http.StatusOK, "list OK")

	router := NewMetricsRouter(
		updateHandler,
		updateHandler,
		getPathHandler,
		getBodyHandler,
		listHandler,
	)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "POST /update/{type}/{name}/{value}",
			method:         http.MethodPost,
			url:            "/update/gauge/testMetric/123",
			expectedStatus: http.StatusOK,
			expectedBody:   "update OK",
		},
		{
			name:           "POST /update/{type}/{name}",
			method:         http.MethodPost,
			url:            "/update/counter/testMetric",
			expectedStatus: http.StatusOK,
			expectedBody:   "update OK",
		},
		{
			name:           "POST /update/",
			method:         http.MethodPost,
			url:            "/update/",
			expectedStatus: http.StatusOK,
			expectedBody:   "update OK",
		},
		{
			name:           "GET /value/{type}/{name}",
			method:         http.MethodGet,
			url:            "/value/counter/testMetric",
			expectedStatus: http.StatusOK,
			expectedBody:   "get path OK",
		},
		{
			name:           "GET /value/{type}",
			method:         http.MethodGet,
			url:            "/value/counter",
			expectedStatus: http.StatusOK,
			expectedBody:   "get path OK",
		},
		{
			name:           "POST /value/",
			method:         http.MethodPost,
			url:            "/value/",
			expectedStatus: http.StatusOK,
			expectedBody:   "get body OK",
		},
		{
			name:           "GET /",
			method:         http.MethodGet,
			url:            "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "list OK",
		},
		{
			name:           "Not Found route",
			method:         http.MethodGet,
			url:            "/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}
