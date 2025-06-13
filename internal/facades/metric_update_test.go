package facades

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateFacade_Update_Success(t *testing.T) {
	var receivedHash string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		receivedHash = r.Header.Get("HashSHA256")
		assert.NotEmpty(t, receivedHash)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New()
	secretKey := "supersecretkey"
	facade := NewMetricUpdateFacade(client, ts.URL, secretKey)

	val := 42.0
	m := types.Metrics{
		ID:    "metric1",
		MType: types.Gauge,
		Value: &val,
	}

	req := []types.Metrics{m, m}

	err := facade.Updates(context.Background(), req)
	assert.NoError(t, err)

	// Считаем хеш от НЕсжатого JSON + secretKey
	jsonData, err := json.Marshal(req)
	assert.NoError(t, err)

	sum := sha256.Sum256(append(jsonData, []byte(secretKey)...))
	expectedHash := hex.EncodeToString(sum[:])

	assert.Equal(t, expectedHash, receivedHash)
}
