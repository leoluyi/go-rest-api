package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	routing "github.com/go-ozzo/ozzo-routing/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// Handler creates a middleware that handles panics and errors encountered during HTTP request processing.
func Handler(logger log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var err error
			defer func() {
				l := logger.With(r.Context())
				if e := recover(); e != nil {
					var ok bool
					if err, ok = e.(error); !ok {
						err = fmt.Errorf("%v", e)
					}

					l.Errorf("recovered from panic (%v): %s", err, debug.Stack())
				}

				if err != nil {
					res := buildErrorResponse(err)
					if res.StatusCode() == http.StatusInternalServerError {
						l.Errorf("encountered internal server error: %v", err)
					}
					w.WriteHeader(res.StatusCode())
					if _, err = w.Write([]byte(res.Message)); err != nil {
						l.Errorf("failed writing error response: %v", err)
					}
					err = nil // return nil because the error is already handled
				}
			}()

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// buildErrorResponse builds an error response from an error.
func buildErrorResponse(err error) ErrorResponse {
	switch err := err.(type) {
	case ErrorResponse:
		return err
	case validation.Errors:
		return InvalidInput(err)
	case routing.HTTPError:
		switch err.StatusCode() {
		case http.StatusNotFound:
			return NotFound("")
		default:
			return ErrorResponse{
				Status:  err.StatusCode(),
				Message: err.Error(),
			}
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return NotFound("")
	}
	return InternalServerError("")
}
