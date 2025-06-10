package facades

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateFacade_Update_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New()
	facade := NewMetricUpdateFacade(client, ts.URL, "/update")

	req := types.MetricsUpdatePathRequest{
		ID:    "metric1",
		MType: "gauge",
		Value: "42",
	}

	err := facade.Update(context.Background(), req)
	assert.NoError(t, err)
}

func TestMetricUpdateFacade_Update_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	client := resty.New()
	facade := NewMetricUpdateFacade(client, ts.URL, "/update")

	req := types.MetricsUpdatePathRequest{
		ID:    "metric1",
		MType: "counter",
		Value: "10",
	}

	err := facade.Update(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server returned status 500")
}

func TestMetricUpdateFacade_Update_ContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer ts.Close()

	client := resty.New()
	facade := NewMetricUpdateFacade(client, ts.URL, "/update")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := types.MetricsUpdatePathRequest{
		ID:    "metric1",
		MType: "gauge",
		Value: "100",
	}

	err := facade.Update(ctx, req)
	assert.Error(t, err)
}

func TestMetricUpdateFacade_AddsHTTPPrefix(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	addr := ts.URL
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")

	client := resty.New()
	facade := NewMetricUpdateFacade(client, addr, "/update")

	req := types.MetricsUpdatePathRequest{
		ID:    "metric1",
		MType: "gauge",
		Value: "42",
	}

	err := facade.Update(context.Background(), req)
	assert.NoError(t, err)
}

func TestMetricUpdateFacade_Update_InvalidServerAddr(t *testing.T) {
	client := resty.New()

	facade := NewMetricUpdateFacade(client, "http://%41:8080", "/update")

	req := types.MetricsUpdatePathRequest{
		ID:    "metric1",
		MType: "gauge",
		Value: "42",
	}

	err := facade.Update(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server address")
}
