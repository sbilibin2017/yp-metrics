package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "PGError with connection code (08006)",
			err:      &pgconn.PgError{Code: "08006"},
			expected: true,
		},
		{
			name:     "PGError with non-connection code (23505)",
			err:      &pgconn.PgError{Code: "23505"},
			expected: false,
		},
		{
			name:     "Non-PgError",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "PGError with short code",
			err:      &pgconn.PgError{Code: "0"},
			expected: false,
		},
		{
			name:     "PGError with correct prefix (08...)",
			err:      &pgconn.PgError{Code: "08ABC"},
			expected: true,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDBConnectionError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsFileLockError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "os.PathError with EAGAIN",
			err: &os.PathError{
				Op:   "open",
				Path: "/tmp/file",
				Err:  syscall.EAGAIN,
			},
			want: true,
		},
		{
			name: "os.PathError with EACCES",
			err: &os.PathError{
				Op:   "open",
				Path: "/tmp/file",
				Err:  syscall.EACCES,
			},
			want: true,
		},
		{
			name: "os.PathError with different errno",
			err: &os.PathError{
				Op:   "open",
				Path: "/tmp/file",
				Err:  syscall.EPERM,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFileLockError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDBRetryMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		handlerFunc     func(http.ResponseWriter, *http.Request, int)
		expectedCode    int
		expectedRetries int
	}{
		{
			name: "Success on first try",
			handlerFunc: func(w http.ResponseWriter, r *http.Request, call int) {
				w.WriteHeader(http.StatusOK)
			},
			expectedCode:    http.StatusOK,
			expectedRetries: 1,
		},
		{
			name: "Non-connection error stops retrying",
			handlerFunc: func(w http.ResponseWriter, r *http.Request, call int) {
				wrw := w.(*responseWriterWithError)
				wrw.err = errors.New("some other db error") // not pgconn.PgError
			},
			expectedCode:    http.StatusOK, // middleware does nothing, handler just returns
			expectedRetries: 1,
		},
		{
			name: "Connection error causes retries then gives up",
			handlerFunc: func(w http.ResponseWriter, r *http.Request, call int) {
				wrw := w.(*responseWriterWithError)
				wrw.err = &pgconn.PgError{Code: "08006"} // will be retried
			},
			expectedCode:    http.StatusInternalServerError,
			expectedRetries: maxAttempts,
		},
		{
			name: "Connection error succeeds on third attempt",
			handlerFunc: func(w http.ResponseWriter, r *http.Request, call int) {
				wrw := w.(*responseWriterWithError)
				if call < 3 {
					wrw.err = &pgconn.PgError{Code: "08006"}
				} else {
					wrw.err = nil
					w.WriteHeader(http.StatusOK)
				}
			},
			expectedCode:    http.StatusOK,
			expectedRetries: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				tt.handlerFunc(w, r, callCount)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			recorder := httptest.NewRecorder()

			rw := &responseWriterWithError{ResponseWriter: recorder}
			RetryMiddleware(handler).ServeHTTP(rw, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)
			assert.Equal(t, tt.expectedRetries, callCount)
		})
	}
}
