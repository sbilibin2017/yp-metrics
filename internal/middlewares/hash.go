package middlewares

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

const hashHeader = "HashSHA256"

func HashMiddleware(secretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if secretKey == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			hash := sha256.Sum256(append(bodyBytes, []byte(secretKey)...))
			expectedHash := hex.EncodeToString(hash[:])

			receivedHash := r.Header.Get(hashHeader)
			if receivedHash == "" || receivedHash != expectedHash {
				http.Error(w, "Invalid hash", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			rw := &responseWriterWithHash{
				ResponseWriter: w,
				secretKey:      secretKey,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(rw, r)

			hashResp := sha256.Sum256(append(rw.body.Bytes(), []byte(secretKey)...))
			respHash := hex.EncodeToString(hashResp[:])

			rw.Header().Set(hashHeader, respHash)
		})
	}
}

type responseWriterWithHash struct {
	http.ResponseWriter
	secretKey string
	body      *bytes.Buffer
}

func (rw *responseWriterWithHash) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}
