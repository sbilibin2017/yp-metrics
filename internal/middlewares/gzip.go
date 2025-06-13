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
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			defer r.Body.Close()
			r.Body = io.NopCloser(gzReader)
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzw := gzip.NewWriter(w)
		defer gzw.Close()

		gzrw := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gzw,
			headerWritten:  false,
		}

		next.ServeHTTP(gzrw, r)

		if !gzrw.headerWritten {
			gzrw.WriteHeader(http.StatusOK)
		}
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer        io.Writer
	headerWritten bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.headerWritten = true

	contentType := w.Header().Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
	} else {
		w.Writer = w.ResponseWriter
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)
}
