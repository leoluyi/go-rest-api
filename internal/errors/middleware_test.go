package errors

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	logger, entries := log.NewForTest()

	t.Run("normal processing", func(t *testing.T) {
		entries.TakeAll()
		handler := Handler(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest("GET", "http://127.0.0.1/users", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		assert.Zero(t, entries.Len())
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("panic processing", func(t *testing.T) {
		entries.TakeAll()
		handler := Handler(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("xyz")
		}))
		req := httptest.NewRequest("GET", "http://127.0.0.1/users", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.GreaterOrEqual(t, entries.Len(), 1)
	})
}

func Test_buildErrorResponse(t *testing.T) {
	res := NotFound("")
	assert.Equal(t, res, buildErrorResponse(res))

	res = buildErrorResponse(sql.ErrNoRows)
	assert.Equal(t, http.StatusNotFound, res.Status)

	v := validator.New()
	type testStruct struct {
		Name string `validate:"required"`
	}
	err := v.Struct(testStruct{})
	var errs validator.ValidationErrors
	assert.ErrorAs(t, err, &errs)
	res = buildErrorResponse(errs)
	assert.Equal(t, http.StatusBadRequest, res.Status)

	res = buildErrorResponse(fmt.Errorf("test"))
	assert.Equal(t, http.StatusInternalServerError, res.Status)
}
