package auth

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth/v5"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/errors"
)

var tokenAuth *jwtauth.JWTAuth

// Handler returns a JWT-based authentication middleware.
func Handler(JWTSigningKey string) func(http.Handler) http.Handler {
	tokenAuth = jwtauth.New("HS256", []byte(JWTSigningKey), nil)
	return jwtauth.Verifier(tokenAuth)
}

// handleToken stores the user identity in the request context so that it can be accessed elsewhere.
func handleToken(c *routing.Context, token *jwt.Token) error {
	ctx := WithUser(
		c.Request.Context(),
		token.Claims.(jwt.MapClaims)["id"].(string),
		token.Claims.(jwt.MapClaims)["name"].(string),
	)
	c.Request = c.Request.WithContext(ctx)
	return nil
}

// WithUser returns a context that contains the user identity from the given JWT.
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

// MockAuthHandler creates a mock authentication middleware for testing purpose.
// If the request contains an Authorization header whose value is "TEST", then
// it considers the user is authenticated as "Tester" whose ID is "100".
// It fails the authentication otherwise.
func MockAuthHandler(c *routing.Context) error {
	if c.Request.Header.Get("Authorization") != "TEST" {
		return errors.Unauthorized("")
	}
	ctx := WithUser(c.Request.Context(), "100", "Tester")
	c.Request = c.Request.WithContext(ctx)
	return nil
}

// MockAuthHeader returns an HTTP header that can pass the authentication check by MockAuthHandler.
func MockAuthHeader() http.Header {
	header := http.Header{}
	header.Add("Authorization", "TEST")
	return header
}
