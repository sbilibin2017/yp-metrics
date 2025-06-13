package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testKey = "supersecret"
const testHeader = "HashSHA256"

func TestHashMiddleware_ValidHash(t *testing.T) {
	body := []byte(`{"metric":"cpu","value":0.42}`)
	hash := computeHMAC(body, []byte(testKey))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(testHeader, hash)

	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := HashMiddleware(testHeader, testKey)
	middleware(handler).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	respHash := resp.Header.Get(testHeader)

	// Проверяем, что тело ответа корректное и хеш ответа правильный
	respBody, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "OK", string(respBody))

	expectedRespHash := computeHMAC([]byte("OK"), []byte(testKey))
	assert.Equal(t, expectedRespHash, respHash)
}

func TestHashMiddleware_InvalidHash(t *testing.T) {
	body := []byte(`{"metric":"memory","value":99}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(testHeader, "invalidhash")

	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Should not reach"))
	})

	middleware := HashMiddleware(testHeader, testKey)
	middleware(handler).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	// Проверяем, что сервер вернул 400 при неверном хеше
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHashMiddleware_NoKey(t *testing.T) {
	body := []byte(`test-no-key`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("No key check"))
	})

	middleware := HashMiddleware(testHeader, "") // Ключ не задан
	middleware(handler).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	// Проверка, что всё отработало без проверки подписи
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "No key check", string(respBody))
	assert.Equal(t, "", resp.Header.Get(testHeader)) // Заголовок подписи не добавляется
}
