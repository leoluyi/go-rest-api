package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/errors"
)

type contextKey int

const userKey contextKey = iota

var tokenAuth *jwtauth.JWTAuth

// Handler returns a JWT-based authentication middleware.
// It combines token verification and authentication: requests with a missing
// or invalid token are rejected with 401.
func Handler(JWTSigningKey string) func(http.Handler) http.Handler {
	tokenAuth = jwtauth.New("HS256", []byte(JWTSigningKey), nil)
	verifier := jwtauth.Verifier(tokenAuth)
	return func(next http.Handler) http.Handler {
		return verifier(jwtauth.Authenticator(next))
	}
}

// WithUser returns a context that contains the user identity.
func WithUser(ctx context.Context, id, name string) context.Context {
	return context.WithValue(ctx, userKey, entity.User{ID: id, Name: name})
}

// CurrentUser returns the user identity from the given context.
// Nil is returned if no user identity is found in the context.
func CurrentUser(ctx context.Context) Identity {
	if user, ok := ctx.Value(userKey).(entity.User); ok {
		return user
	}
	return nil
}

// MockAuthHandler is an authentication middleware for testing.
// Requests with Authorization header value "TEST" are authenticated as user "100"/"Tester".
// All other requests receive a 401 response.
func MockAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "TEST" {
			errors.RespondWithError(w, errors.Unauthorized(""))
			return
		}
		ctx := WithUser(r.Context(), "100", "Tester")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MockAuthHeader returns an HTTP header that passes the MockAuthHandler check.
func MockAuthHeader() http.Header {
	header := http.Header{}
	header.Add("Authorization", "TEST")
	return header
}
