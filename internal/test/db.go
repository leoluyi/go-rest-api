package test

import (
	"context"
	"path"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // initialize postgresql for test
	"github.com/leoluyi/go-api-template/internal/config"
	"github.com/leoluyi/go-api-template/pkg/dbcontext"
	"github.com/leoluyi/go-api-template/pkg/log"
)

var db *dbcontext.DB

// DB returns the database connection for testing purpose.
func DB(t *testing.T) *dbcontext.DB {
	if db != nil {
		return db
	}
	logger, _ := log.NewForTest()
	dir := getSourcePath()
	cfg, err := config.Load(dir+"/../../config/local.yml", logger)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	dbc, err := sqlx.Open("postgres", cfg.DSN)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	db = dbcontext.New(dbc)
	return db
}

// ResetTables truncates all data in the specified tables.
func ResetTables(t *testing.T, db *dbcontext.DB, tables ...string) {
	for _, table := range tables {
		if _, err := db.DB().ExecContext(context.Background(), "TRUNCATE "+table); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

// getSourcePath returns the directory containing the source code that is calling this function.
func getSourcePath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
