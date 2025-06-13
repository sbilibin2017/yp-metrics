package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testKey    = "supersecretkey"
	testHeader = "HashSHA256"
)

func computeTestHMAC(body []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestHashMiddleware_ValidHash(t *testing.T) {
	body := []byte(`{"metric":"value"}`)
	expectedHash := computeTestHMAC(body, testKey)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(testHeader, expectedHash)

	rr := httptest.NewRecorder()

	// dummy handler that returns a known response
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`OK`))
	})

	middleware := HashMiddleware(testKey, testHeader)
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())

	respHash := rr.Header().Get(testHeader)
	expectedRespHash := computeTestHMAC([]byte("OK"), testKey)
	assert.Equal(t, expectedRespHash, respHash)
}

func TestHashMiddleware_InvalidHash(t *testing.T) {
	body := []byte(`{"metric":"value"}`)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(testHeader, "invalidhash")

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called on invalid hash")
	})

	middleware := HashMiddleware(testKey, testHeader)
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHashMiddleware_NoKey(t *testing.T) {
	body := []byte(`{"metric":"value"}`)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	called := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	middleware := HashMiddleware("", testHeader)
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusTeapot, rr.Code)
}
