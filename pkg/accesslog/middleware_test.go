package accesslog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	logger, entries := log.NewForTest()
	handler := Handler(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://127.0.0.1/users", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	assert.Equal(t, 1, entries.Len())
	assert.Equal(t, "GET /users HTTP/1.1 200 0", entries.All()[0].Message)
}

func TestHandler_bytesWritten(t *testing.T) {
	logger, entries := log.NewForTest()
	body := []byte("hello world")
	handler := Handler(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))

	req := httptest.NewRequest("GET", "http://127.0.0.1/users", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	assert.Equal(t, 1, entries.Len())
	// bytesWritten should reflect the actual body size written.
	assert.Equal(t, "GET /users HTTP/1.1 200 11", entries.All()[0].Message)
}

func TestHandler_propagatesRequestID(t *testing.T) {
	logger, _ := log.NewForTest()
	handler := Handler(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://127.0.0.1/users", nil)
	req.Header.Set("X-Request-ID", "req-abc-123")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	assert.Equal(t, "req-abc-123", res.Header().Get("X-Request-ID"))
}
