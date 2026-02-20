package accesslog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qiangxue/go-rest-api/pkg/log"
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
