// Package dbcontext provides DB transaction support for transactions that span method calls of multiple
// repositories and services.
package dbcontext

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
)

// DB represents a DB connection that can be used to run SQL queries.
type DB struct {
	db *sqlx.DB
}

type contextKey int

const (
	txKey contextKey = iota
)

// New returns a new DB connection that wraps the given sqlx.DB instance.
func New(db *sqlx.DB) *DB {
	return &DB{db}
}

// DB returns the sqlx.DB wrapped by this object.
func (db *DB) DB() *sqlx.DB {
	return db.db
}

// With returns a sqlx.ExtContext that can be used to run SQL queries.
// With returns the transaction if one is found in the given context.
// Otherwise it returns the underlying DB connection.
func (db *DB) With(ctx context.Context) sqlx.ExtContext {
	if tx, ok := ctx.Value(txKey).(*sqlx.Tx); ok {
		return tx
	}
	return db.db
}

// Transactional starts a transaction and calls the given function with a context storing the transaction.
// The transaction associated with the context can be accessed via With().
// The transaction is committed if the function returns nil, rolled back otherwise.
func (db *DB) Transactional(ctx context.Context, f func(ctx context.Context) error) (err error) {
	tx, err := db.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				fmt.Fprintf(os.Stderr, "dbcontext: failed to rollback transaction: %v\n", rbErr)
			}
		} else {
			err = tx.Commit()
		}
	}()
	err = f(context.WithValue(ctx, txKey, tx))
	return err
}

// TransactionHandler returns a middleware that wraps each request in a database transaction.
// The transaction is committed when the handler returns normally, and rolled back on panic.
// The transaction can be accessed via With() using the request context.
func (db *DB) TransactionHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tx, err := db.db.BeginTxx(r.Context(), nil)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			panicked := true
			defer func() {
				if panicked {
					if rbErr := tx.Rollback(); rbErr != nil {
						fmt.Fprintf(os.Stderr, "dbcontext: failed to rollback transaction: %v\n", rbErr)
					}
				} else {
					if err := tx.Commit(); err != nil {
						fmt.Fprintf(os.Stderr, "dbcontext: failed to commit transaction: %v\n", err)
					}
				}
			}()
			ctx := context.WithValue(r.Context(), txKey, tx)
			next.ServeHTTP(w, r.WithContext(ctx))
			panicked = false
		})
	}
}
