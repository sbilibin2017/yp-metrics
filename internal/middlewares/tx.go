package middlewares

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func TxMiddleware(db *sqlx.DB, txSetter func(ctx context.Context, tx *sqlx.Tx) context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if db == nil {
				next.ServeHTTP(w, r)
				return
			}

			tx, err := db.Beginx()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := txSetter(r.Context(), tx)
			r = r.WithContext(ctx)

			defer func() {
				if rec := recover(); rec != nil {
					tx.Rollback()
					panic(rec)
				}
			}()

			next.ServeHTTP(w, r)

			if err := tx.Commit(); err != nil {
				tx.Rollback()
			}
		})
	}
}
