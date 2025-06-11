package middlewares

import (
	"net/http"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
	"go.uber.org/zap"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger.Log.Desugar().Info("Request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Duration("duration", duration),
		)

		logger.Log.Desugar().Info("Response",
			zap.Int("status", rw.statusCode),
			zap.Int("response_size", rw.writtenSize),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	writtenSize int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.writtenSize += n
	return n, err
}
