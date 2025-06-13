package middlewares

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func HashMiddleware(secretKey string, hashHeader string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if secretKey == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body.Close()

			hash := sha256.Sum256(append(bodyBytes, []byte(secretKey)...))
			expectedHash := hex.EncodeToString(hash[:])

			receivedHash := r.Header.Get(hashHeader)
			if receivedHash != expectedHash {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			rw := &responseWriterWithHash{
				ResponseWriter: w,
				secretKey:      secretKey,
				body:           &bytes.Buffer{},
				hashHeader:     hashHeader,
			}

			next.ServeHTTP(rw, r)

			rw.Header().Set(hashHeader, rw.computeResponseHash())
		})
	}
}

type responseWriterWithHash struct {
	http.ResponseWriter
	secretKey  string
	body       *bytes.Buffer
	hashHeader string
}

func (rw *responseWriterWithHash) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriterWithHash) computeResponseHash() string {
	hashResp := sha256.Sum256(append(rw.body.Bytes(), []byte(rw.secretKey)...))
	return hex.EncodeToString(hashResp[:])
}
