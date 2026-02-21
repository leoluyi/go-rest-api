package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	h := Handler()
	assert.NotNil(t, h)

	req := httptest.NewRequest("GET", "/metrics", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Header().Get("Content-Type"), "text/plain")
}

func TestMiddleware_recordsMetrics(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Middleware)
	r.Get("/albums/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	before := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/albums/{id}", "200"))

	req := httptest.NewRequest("GET", "/albums/abc-123", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	after := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/albums/{id}", "200"))
	assert.Equal(t, before+1, after, "request counter should be incremented by 1")
}

func TestMiddleware_usesRoutePattern(t *testing.T) {
	// Verify that the chi route pattern (/albums/{id}) is used as the label,
	// not the concrete URL path (/albums/specific-id-123), avoiding high cardinality.
	r := chi.NewRouter()
	r.Use(Middleware)
	r.Get("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	beforePattern := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/items/{id}", "404"))

	req := httptest.NewRequest("GET", "/items/specific-id-123", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	afterPattern := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/items/{id}", "404"))
	afterRaw := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/items/specific-id-123", "404"))

	assert.Equal(t, beforePattern+1, afterPattern, "counter should use route pattern label")
	assert.Equal(t, float64(0), afterRaw, "counter should not use raw URL as label")
}

func TestMiddleware_fallsBackToURLPath(t *testing.T) {
	// When no chi context is present (e.g. middleware used standalone),
	// the raw URL path is used as the label.
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler = Middleware(handler)

	before := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/standalone", "200"))

	req := httptest.NewRequest("GET", "/standalone", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	after := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/standalone", "200"))
	assert.Equal(t, before+1, after)
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, rw.status, "internal status field should be updated")
	assert.Equal(t, http.StatusCreated, rec.Code, "underlying ResponseWriter should receive the status")
}
