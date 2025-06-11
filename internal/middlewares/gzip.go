package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			r.Body = io.NopCloser(gzReader)
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzw := gzip.NewWriter(w)
			defer gzw.Close()

			w.Header().Set("Content-Encoding", "gzip")
			gzwResponseWriter := &gzipResponseWriter{Writer: gzw, ResponseWriter: w}
			next.ServeHTTP(gzwResponseWriter, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
