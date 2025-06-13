package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func HashMiddleware(hashKey string, hashHeader string) func(http.Handler) http.Handler {
	if hashKey == "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body.Close()

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			expectedHash := computeHMAC(bodyBytes, hashKey)
			receivedHash := r.Header.Get(hashHeader)
			if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			recorder := &responseCaptureWriter{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				header:         http.Header{},
			}

			next.ServeHTTP(recorder, r)

			for k, vv := range recorder.header {
				for _, v := range vv {
					w.Header().Add(k, v)
				}
			}

			w.Header().Set(hashHeader, computeHMAC(recorder.body.Bytes(), hashKey))

			if recorder.statusCode == 0 {
				recorder.statusCode = http.StatusOK
			}

			w.WriteHeader(recorder.statusCode)
			_, _ = w.Write(recorder.body.Bytes())
		})
	}
}

type responseCaptureWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	header     http.Header
}

func (w *responseCaptureWriter) Header() http.Header {
	return w.header
}

func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *responseCaptureWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func computeHMAC(body []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
