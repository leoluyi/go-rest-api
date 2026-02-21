package test

import (
	"context"
	"os"
	"path"
	"runtime"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // initialize postgresql for test
	"github.com/leoluyi/go-api-template/internal/config"
	"github.com/leoluyi/go-api-template/pkg/dbcontext"
)

var (
	testDB   *dbcontext.DB
	testDBMu sync.Once
	testDBErr error
)

// DB returns the database connection for testing purpose.
// If APP_DSN is set, it is used directly; otherwise the local config file is loaded.
// The connection is established once and reused across tests.
func DB(t *testing.T) *dbcontext.DB {
	t.Helper()
	testDBMu.Do(func() {
		dsn := os.Getenv("APP_DSN")
		if dsn == "" {
			dir := getSourcePath()
			cfg, err := config.Load(dir + "/../../config/local.yml")
			if err != nil {
				testDBErr = err
				return
			}
			dsn = cfg.DSN
		}
		dbc, err := sqlx.Open("postgres", dsn)
		if err != nil {
			testDBErr = err
			return
		}
		testDB = dbcontext.New(dbc)
	})
	if testDBErr != nil {
		t.Error(testDBErr)
		t.FailNow()
	}
	return testDB
}

// ResetTables truncates all data in the specified tables.
func ResetTables(t *testing.T, db *dbcontext.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if _, err := db.DB().ExecContext(context.Background(), "TRUNCATE "+table+" CASCADE"); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

// getSourcePath returns the directory containing the source file that calls this function.
func getSourcePath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
