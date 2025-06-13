package facades

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func computeHMAC(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func TestMetricUpdateFacade_Update_Success(t *testing.T) {
	val := 42.0
	m := types.Metrics{
		ID:    "metric1",
		MType: types.Gauge,
		Value: &val,
	}
	req := []types.Metrics{m, m}

	// Marshal body to compute expected HMAC
	client := resty.New()
	jsonBytes, err := client.JSONMarshal(req)
	assert.NoError(t, err)
	hashKey := "secret"
	hashHeader := "HashSHA256"
	expectedHMAC := computeHMAC(jsonBytes, hashKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		// Проверяем, что заголовок с хешем присутствует и корректен
		assert.Equal(t, expectedHMAC, r.Header.Get(hashHeader))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	facade := NewMetricUpdateFacade(client, ts.URL, hashKey, hashHeader)
	err = facade.Updates(context.Background(), req)
	assert.NoError(t, err)
}

func TestMetricUpdateFacade_Update_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	client := resty.New()
	facade := NewMetricUpdateFacade(client, ts.URL, "", "")

	val := int64(10)
	m := types.Metrics{
		ID:    "metric1",
		MType: types.Counter,
		Delta: &val,
	}
	req := []types.Metrics{m, m}

	err := facade.Updates(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server returned status 500")
}

func TestMetricUpdateFacade_Update_ContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer ts.Close()

	client := resty.New()
	facade := NewMetricUpdateFacade(client, ts.URL, "", "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	val := int64(10)
	m := types.Metrics{
		ID:    "metric1",
		MType: types.Counter,
		Delta: &val,
	}
	req := []types.Metrics{m, m}

	err := facade.Updates(ctx, req)
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
	facade := NewMetricUpdateFacade(client, addr, "", "")

	val := int64(10)
	m := types.Metrics{
		ID:    "metric1",
		MType: types.Counter,
		Delta: &val,
	}
	req := []types.Metrics{m, m}

	err := facade.Updates(context.Background(), req)
	assert.NoError(t, err)
}
