package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/validator/v10"
	"github.com/leoluyi/go-api-template/pkg/log"
)

// Handler creates a middleware that handles panics encountered during HTTP request processing.
func Handler(logger log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if e := recover(); e != nil {
					l := logger.With(r.Context())
					var err error
					var ok bool
					if err, ok = e.(error); !ok {
						err = fmt.Errorf("%v", e)
					}
					l.Errorf("recovered from panic (%v): %s", err, debug.Stack())
					res := buildErrorResponse(err)
					if res.StatusCode() == http.StatusInternalServerError {
						l.Errorf("encountered internal server error: %v", err)
					}
					RespondJSON(w, res.Status, res)
				}
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// buildErrorResponse builds an ErrorResponse from an error.
func buildErrorResponse(err error) ErrorResponse {
	switch err := err.(type) {
	case ErrorResponse:
		return err
	case validator.ValidationErrors:
		return InvalidInput(err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound("")
	}
	return InternalServerError("")
}
