// Package accesslog provides a middleware that records every RESTful API call in a log message.
package accesslog

import (
	"net/http"
	"time"

	"github.com/go-ozzo/ozzo-routing/v2/access"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// Handler returns a middleware that records an access log message for every HTTP request being processed.
// https://github.com/go-chi/httplog/blob/master/httplog.go#L44
func Handler(logger log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &access.LogResponseWriter{
				ResponseWriter: w,
				Status:         http.StatusOK,
			}

			// associate request ID and session ID with the request context
			// so that they can be added to the log messages
			ctx := r.Context()
			ctx = log.WithRequest(ctx, r)

			next.ServeHTTP(rw, r.WithContext(ctx))

			// generate an access log message
			logger.With(ctx, "duration", time.Since(start).Milliseconds(), "status", rw.Status).
				Infof("%s %s %s %d %d", r.Method, r.URL.Path, r.Proto, rw.Status, rw.BytesWritten)
		}
		return http.HandlerFunc(fn)
	}
}
