package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentUser(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, CurrentUser(ctx))
	ctx = WithUser(ctx, "100", "test")
	identity := CurrentUser(ctx)
	if assert.NotNil(t, identity) {
		assert.Equal(t, "100", identity.GetID())
		assert.Equal(t, "test", identity.GetName())
	}
}

func TestHandler(t *testing.T) {
	assert.NotNil(t, Handler("test"))
}

func TestMockAuthHandler(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.NotNil(t, CurrentUser(r.Context()))
	})

	// unauthorized request
	req := httptest.NewRequest("GET", "http://example.com", nil)
	res := httptest.NewRecorder()
	MockAuthHandler(next).ServeHTTP(res, req)
	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, res.Code)

	// authorized request
	req = httptest.NewRequest("GET", "http://example.com", nil)
	req.Header = MockAuthHeader()
	res = httptest.NewRecorder()
	MockAuthHandler(next).ServeHTTP(res, req)
	assert.True(t, called)
}
