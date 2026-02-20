package healthcheck

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/leoluyi/go-api-template/internal/test"
	"github.com/leoluyi/go-api-template/pkg/log"
)

type mockDB struct{ err error }

func (m mockDB) PingContext(_ context.Context) error { return m.err }

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()

	router := test.MockRouter(logger)
	RegisterHandlers(router, "0.9.0", mockDB{})
	test.Endpoint(t, router, test.APITestCase{
		Name:         "healthy",
		Method:       "GET",
		URL:          "/healthcheck",
		WantStatus:   http.StatusOK,
		WantResponse: `{"status":"ok","version":"0.9.0","db":"ok"}`,
	})

	router2 := test.MockRouter(logger)
	RegisterHandlers(router2, "0.9.0", mockDB{err: errors.New("connection refused")})
	test.Endpoint(t, router2, test.APITestCase{
		Name:         "degraded",
		Method:       "GET",
		URL:          "/healthcheck",
		WantStatus:   http.StatusServiceUnavailable,
		WantResponse: `*"status":"degraded"*`,
	})
}
