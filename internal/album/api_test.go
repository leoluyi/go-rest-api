package album

import (
	"net/http"
	"testing"
	"time"

	"github.com/qiangxue/go-rest-api/internal/auth"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/test"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mockRepository{items: []entity.Album{
		{ID: "123", Name: "album123", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	RegisterHandlers(router, NewService(repo, logger), auth.MockAuthHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{Name: "get all", Method: "GET", URL: "/albums", WantStatus: http.StatusOK, WantResponse: `*"total_count":1*`},
		{Name: "get 123", Method: "GET", URL: "/albums/123", WantStatus: http.StatusOK, WantResponse: `*album123*`},
		{Name: "get unknown", Method: "GET", URL: "/albums/1234", WantStatus: http.StatusNotFound},
		{Name: "create ok", Method: "POST", URL: "/albums", Body: `{"name":"test"}`, Header: header, WantStatus: http.StatusCreated, WantResponse: "*test*"},
		{Name: "create ok count", Method: "GET", URL: "/albums", WantStatus: http.StatusOK, WantResponse: `*"total_count":2*`},
		{Name: "create auth error", Method: "POST", URL: "/albums", Body: `{"name":"test"}`, WantStatus: http.StatusUnauthorized},
		{Name: "create input error", Method: "POST", URL: "/albums", Body: `"name":"test"}`, Header: header, WantStatus: http.StatusBadRequest},
		{Name: "update ok", Method: "PUT", URL: "/albums/123", Body: `{"name":"albumxyz"}`, Header: header, WantStatus: http.StatusOK, WantResponse: "*albumxyz*"},
		{Name: "update verify", Method: "GET", URL: "/albums/123", WantStatus: http.StatusOK, WantResponse: `*albumxyz*`},
		{Name: "update auth error", Method: "PUT", URL: "/albums/123", Body: `{"name":"albumxyz"}`, WantStatus: http.StatusUnauthorized},
		{Name: "update input error", Method: "PUT", URL: "/albums/123", Body: `"name":"albumxyz"}`, Header: header, WantStatus: http.StatusBadRequest},
		{Name: "delete ok", Method: "DELETE", URL: "/albums/123", Header: header, WantStatus: http.StatusOK, WantResponse: "*albumxyz*"},
		{Name: "delete verify", Method: "DELETE", URL: "/albums/123", Header: header, WantStatus: http.StatusNotFound},
		{Name: "delete auth error", Method: "DELETE", URL: "/albums/123", WantStatus: http.StatusUnauthorized},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
