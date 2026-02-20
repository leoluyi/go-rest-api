package dbcontext

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // initialize postgresql for test
	"github.com/stretchr/testify/assert"
)

const DSN = "postgres://127.0.0.1/go_restful?sslmode=disable&user=postgres&password=postgres"

func TestNew(t *testing.T) {
	runDBTest(t, func(db *sqlx.DB) {
		dbc := New(db)
		assert.NotNil(t, dbc)
		assert.Equal(t, db, dbc.DB())
	})
}

func TestDB_Transactional(t *testing.T) {
	runDBTest(t, func(db *sqlx.DB) {
		assert.Zero(t, runCountQuery(t, db))
		dbc := New(db)

		// successful transaction
		err := dbc.Transactional(context.Background(), func(ctx context.Context) error {
			_, err := dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "1", "name1")
			assert.Nil(t, err)
			_, err = dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "2", "name2")
			assert.Nil(t, err)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 2, runCountQuery(t, db))

		// failed transaction
		err = dbc.Transactional(context.Background(), func(ctx context.Context) error {
			_, err := dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "3", "name3")
			assert.Nil(t, err)
			_, err = dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "4", "name4")
			assert.Nil(t, err)
			return sql.ErrNoRows
		})
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Equal(t, 2, runCountQuery(t, db))

		// queries made outside of the transaction are not rolled back
		err = dbc.Transactional(context.Background(), func(ctx context.Context) error {
			_, err := dbc.With(context.Background()).ExecContext(context.Background(), "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "3", "name3")
			assert.Nil(t, err)
			_, err = dbc.With(context.Background()).ExecContext(context.Background(), "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "4", "name4")
			assert.Nil(t, err)
			return sql.ErrNoRows
		})
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Equal(t, 4, runCountQuery(t, db))
	})
}

func TestDB_TransactionHandler(t *testing.T) {
	runDBTest(t, func(db *sqlx.DB) {
		assert.Zero(t, runCountQuery(t, db))
		dbc := New(db)

		// successful transaction
		{
			req := httptest.NewRequest("GET", "http://127.0.0.1/users", nil)
			res := httptest.NewRecorder()
			handler := dbc.TransactionHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				_, err := dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "1", "name1")
				assert.Nil(t, err)
				_, err = dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "2", "name2")
				assert.Nil(t, err)
			}))
			handler.ServeHTTP(res, req)
			assert.Equal(t, 2, runCountQuery(t, db))
		}

		// failed transaction (panic causes rollback)
		{
			req := httptest.NewRequest("GET", "http://127.0.0.1/users", nil)
			res := httptest.NewRecorder()
			handler := dbc.TransactionHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				_, err := dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "3", "name3")
				assert.Nil(t, err)
				_, err = dbc.With(ctx).ExecContext(ctx, "INSERT INTO dbcontexttest (id, name) VALUES ($1, $2)", "4", "name4")
				assert.Nil(t, err)
				panic("simulate error")
			}))
			assert.Panics(t, func() { handler.ServeHTTP(res, req) })
			assert.Equal(t, 2, runCountQuery(t, db))
		}
	})
}

func runDBTest(t *testing.T, f func(db *sqlx.DB)) {
	dsn, ok := os.LookupEnv("APP_DSN")
	if !ok {
		dsn = DSN
	}
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer func() {
		_ = db.Close()
	}()

	sqls := []string{
		"CREATE TABLE IF NOT EXISTS dbcontexttest (id VARCHAR PRIMARY KEY, name VARCHAR)",
		"TRUNCATE dbcontexttest",
	}
	for _, s := range sqls {
		if _, err = db.Exec(s); err != nil {
			t.Errorf("%v with SQL: %s", err, s)
			t.FailNow()
		}
	}

	f(db)
}

func runCountQuery(t *testing.T, db *sqlx.DB) int {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM dbcontexttest")
	assert.Nil(t, err)
	return count
}
