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

	// Тело запроса
	body := []byte(`{"id":"metric1","type":"gauge","value":123.45}`)

	// Вычисляем хеш от тела + ключ
	sum := sha256.Sum256(append(body, []byte(secretKey)...))
	expectedHash := hex.EncodeToString(sum[:])

	// Создаем запрос с заголовком HashSHA256
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("HashSHA256", expectedHash)

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Следующий handler просто возвращает 200 OK и тело
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(body)
		require.NoError(t, err)
	})

	// Оборачиваем middleware
	handler := HashMiddleware(secretKey)(nextHandler)

	handler.ServeHTTP(rr, req)

	// Проверяем, что статус 200
	require.Equal(t, http.StatusOK, rr.Code)

	// Проверяем, что в ответе есть заголовок HashSHA256 и он корректен
	respBody := rr.Body.Bytes()
	sumResp := sha256.Sum256(append(respBody, []byte(secretKey)...))
	expectedRespHash := hex.EncodeToString(sumResp[:])

	require.Equal(t, expectedRespHash, rr.Header().Get("HashSHA256"))

	// Проверяем, что тело совпадает
	require.Equal(t, body, respBody)
}

func TestHashMiddleware_InvalidHash(t *testing.T) {
	secretKey := "testkey"
	body := []byte(`{"id":"metric1","type":"gauge","value":123.45}`)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	// Устанавливаем неправильный хеш
	req.Header.Set("HashSHA256", "invalidhash")

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on invalid hash")
	})

	handler := HashMiddleware(secretKey)(nextHandler)

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "Invalid hash")
}

func TestHashMiddleware_EmptyKey(t *testing.T) {
	// Если ключ пустой, middleware пропускает запрос без проверки
	secretKey := ""

	body := []byte(`{"id":"metric1","type":"gauge","value":123.45}`)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	handler := HashMiddleware(secretKey)(nextHandler)

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "ok", rr.Body.String())
	require.Empty(t, rr.Header().Get("HashSHA256"))
}
