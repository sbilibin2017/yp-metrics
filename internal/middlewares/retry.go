package middlewares

import (
	"errors"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/jackc/pgconn"
)

const maxAttempts = 4

var delays = []time.Duration{0, 1 * time.Second, 3 * time.Second, 5 * time.Second}

var retryErrCheckers = []func(err error) bool{
	isDBConnectionError,
	isFileLockError,
}

func RetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for attempt := 0; attempt < maxAttempts; attempt++ {
			if attempt > 0 {
				time.Sleep(delays[attempt])
			}

			rw := &responseWriterWithError{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			if rw.err == nil {
				return
			}

			var isRetryErr bool
			for _, checker := range retryErrCheckers {
				if checker(rw.err) {
					isRetryErr = true
					break
				}
			}
			if !isRetryErr {
				return
			}
		}

		// Removed http.Error; just set status code
		w.WriteHeader(http.StatusInternalServerError)
	})
}

type responseWriterWithError struct {
	http.ResponseWriter
	err error
}

func (rw *responseWriterWithError) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.err = err
	return n, err
}

func (rw *responseWriterWithError) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
}

func isDBConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08"
}

func isFileLockError(err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		if errors.Is(pathErr.Err, syscall.EAGAIN) || errors.Is(pathErr.Err, syscall.EACCES) {
			return true
		}
	}
	return false
}
