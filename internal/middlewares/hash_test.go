package middlewares

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashMiddleware_ValidHash(t *testing.T) {
	secretKey := "testkey"
	const hashHeader = "HashSHA256"

	body := []byte(`{"id":"metric1","type":"gauge","value":123.45}`)

	sum := sha256.Sum256(append(body, []byte(secretKey)...))
	expectedHash := hex.EncodeToString(sum[:])

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set(hashHeader, expectedHash)

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(body)
		require.NoError(t, err)
	})

	handler := HashMiddleware(secretKey, hashHeader)(nextHandler)

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	respBody := rr.Body.Bytes()
	sumResp := sha256.Sum256(append(respBody, []byte(secretKey)...))
	expectedRespHash := hex.EncodeToString(sumResp[:])

	require.Equal(t, expectedRespHash, rr.Header().Get(hashHeader))
	require.Equal(t, body, respBody)
}
