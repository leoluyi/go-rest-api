package test

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/accesslog"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// MockRouter creates a chi.Router for testing APIs.
func MockRouter(logger log.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		cors.AllowAll().Handler,
	)
	return r
}
