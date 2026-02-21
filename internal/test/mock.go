package test

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/leoluyi/go-api-template/internal/errors"
	"github.com/leoluyi/go-api-template/pkg/accesslog"
	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/leoluyi/go-api-template/pkg/metrics"
)

// MockRouter creates a chi.Router for testing APIs.
// The middleware stack mirrors the production router so that tests catch
// interactions between handlers and middleware.
func MockRouter(logger log.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		metrics.Middleware,
		cors.AllowAll().Handler,
	)
	return r
}
