package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipMiddleware_CompressedRequestAndResponse(t *testing.T) {

	originalBody := `{"message":"Hello, server!"}`
	var compressedRequest bytes.Buffer
	gzw := gzip.NewWriter(&compressedRequest)
	_, err := gzw.Write([]byte(originalBody))
	assert.NoError(t, err)
	gzw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &compressedRequest)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})

	GzipMiddleware(handler).ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))

	respReader, err := gzip.NewReader(recorder.Body)
	assert.NoError(t, err)
	defer respReader.Close()

	uncompressedResp, err := io.ReadAll(respReader)
	assert.NoError(t, err)

	assert.Equal(t, originalBody, string(uncompressedResp))
}

func TestGzipMiddleware_PlainRequestAndResponse(t *testing.T) {
	// Обычный JSON без сжатия
	plainBody := `{"plain":"request"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(plainBody))
	// Нет заголовков Content-Encoding или Accept-Encoding

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})

	GzipMiddleware(handler).ServeHTTP(recorder, req)

	// Проверка: Content-Encoding не должен быть установлен
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Header().Get("Content-Encoding"))

	// Проверка тела ответа
	assert.Equal(t, plainBody, recorder.Body.String())
}
