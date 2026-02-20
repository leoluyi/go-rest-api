// Package accesslog provides a middleware that records every RESTful API call in a log message.
package accesslog

import (
	"net/http"
	"time"

	"github.com/qiangxue/go-rest-api/pkg/log"
)

type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// Handler returns a middleware that records an access log message for every HTTP request being processed.
func Handler(logger log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			ctx := log.WithRequest(r.Context(), r)
			next.ServeHTTP(rw, r.WithContext(ctx))
			logger.With(ctx, "duration", time.Since(start).Milliseconds(), "status", rw.status).
				Infof("%s %s %s %d %d", r.Method, r.URL.Path, r.Proto, rw.status, rw.bytesWritten)
		}
		return http.HandlerFunc(fn)
	}
}
