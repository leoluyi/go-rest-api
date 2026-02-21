package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/leoluyi/go-api-template/internal/entity"
	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestHandlerVerifiesJWT(t *testing.T) {
	const signingKey = "test_signing_key_must_be_32_chars"

	// Generate tokens using the same signing logic as the auth service.
	logger, _ := log.NewForTest()
	svc := service{signingKey, 1, "demo", "pass", logger}

	validToken, err := svc.generateJWT(entity.User{ID: "100", Name: "demo"})
	require.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   "100",
		"name": "demo",
		"exp":  time.Now().Add(-time.Hour).Unix(),
	}).SignedString([]byte(signingKey))
	require.NoError(t, err)

	// Mount Handler on a chi router protecting a test endpoint.
	r := chi.NewRouter()
	r.Use(Handler(signingKey))
	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"valid token", "Bearer " + validToken, http.StatusOK},
		{"missing token", "", http.StatusUnauthorized},
		{"malformed token", "Bearer not-a-valid-jwt", http.StatusUnauthorized},
		{"expired token", "Bearer " + expiredToken, http.StatusUnauthorized},
		{"wrong signature", "Bearer " + validToken + "tampered", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			assert.Equal(t, tt.wantStatus, res.Code)
		})
	}
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
