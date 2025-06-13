package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func HashMiddleware(hashHeader string, hashKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hashKey != "" {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body.Close()
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				expectedHash := computeHMAC(bodyBytes, []byte(hashKey))
				receivedHash := r.Header.Get(hashHeader)

				if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}

			if hashKey != "" {
				buffer := &bytes.Buffer{}
				hashWriter := &responseWriterWithBody{ResponseWriter: w, body: buffer}
				next.ServeHTTP(hashWriter, r)

				w.Header().Set(hashHeader, computeHMAC(buffer.Bytes(), []byte(hashKey)))
				w.WriteHeader(hashWriter.statusCode)
				w.Write(buffer.Bytes())
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func computeHMAC(data, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

type responseWriterWithBody struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriterWithBody) WriteHeader(code int) {
	w.statusCode = code
}

func (w *responseWriterWithBody) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
