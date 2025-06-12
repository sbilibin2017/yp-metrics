package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/yp-metrics/internal/contexts"
)

// simple txSetter to put tx into context
func testTxSetter(ctx context.Context, tx *sqlx.Tx) context.Context {
	return contexts.SetTxToContext(ctx, tx)
}

func TestTxMiddleware_SuccessCommit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectCommit()

	called := false
	handler := TxMiddleware(sqlxDB, testTxSetter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		tx := contexts.GetTxFromContext(r.Context())
		require.NotNil(t, tx)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin().WillReturnError(assert.AnError)

	handler := TxMiddleware(sqlxDB, testTxSetter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTxMiddleware_Panic_RollbackAndPanic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectRollback()

	handler := TxMiddleware(sqlxDB, testTxSetter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	assert.Panics(t, func() {
		handler.ServeHTTP(w, req)
	})

	// В этом тесте мы не вызываем w.Result(),
	// но если добавите - обязательно закрывайте Body
	require.NoError(t, mock.ExpectationsWereMet())
}
