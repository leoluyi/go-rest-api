package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	pkglog "github.com/leoluyi/go-api-template/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestErrorResponse_Error(t *testing.T) {
	e := ErrorResponse{
		Message: "abc",
	}
	assert.Equal(t, "abc", e.Error())
}

func TestErrorResponse_StatusCode(t *testing.T) {
	e := ErrorResponse{
		Status: 400,
	}
	assert.Equal(t, 400, e.StatusCode())
}

func TestInternalServerError(t *testing.T) {
	res := InternalServerError("test")
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
	assert.Equal(t, "test", res.Error())
	res = InternalServerError("")
	assert.NotEmpty(t, res.Error())
}

func TestNotFound(t *testing.T) {
	res := NotFound("test")
	assert.Equal(t, http.StatusNotFound, res.StatusCode())
	assert.Equal(t, "test", res.Error())
	res = NotFound("")
	assert.NotEmpty(t, res.Error())
}

func TestUnauthorized(t *testing.T) {
	res := Unauthorized("test")
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode())
	assert.Equal(t, "test", res.Error())
	res = Unauthorized("")
	assert.NotEmpty(t, res.Error())
}

func TestForbidden(t *testing.T) {
	res := Forbidden("test")
	assert.Equal(t, http.StatusForbidden, res.StatusCode())
	assert.Equal(t, "test", res.Error())
	res = Forbidden("")
	assert.NotEmpty(t, res.Error())
}

func TestBadRequest(t *testing.T) {
	res := BadRequest("test")
	assert.Equal(t, http.StatusBadRequest, res.StatusCode())
	assert.Equal(t, "test", res.Error())
	res = BadRequest("")
	assert.NotEmpty(t, res.Error())
}

func TestInvalidInput(t *testing.T) {
	v := validator.New()
	type testStruct struct {
		Name string `validate:"required"`
	}
	err := v.Struct(testStruct{})
	var errs validator.ValidationErrors
	assert.ErrorAs(t, err, &errs)
	res := InvalidInput(errs)
	assert.Equal(t, http.StatusBadRequest, res.Status)
	details, ok := res.Details.([]invalidField)
	assert.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "Name", details[0].Field)
}

func TestRespondJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	RespondJSON(rec, http.StatusCreated, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var body map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "value", body["key"])
}

func TestRespondWithError(t *testing.T) {
	t.Run("ErrorResponse sets correct status and body", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		RespondWithError(rec, req, NotFound("item not found"))

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]interface{}
		require := assert.New(t)
		require.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, float64(http.StatusNotFound), body["status"])
		assert.Equal(t, "item not found", body["message"])
	})

	t.Run("generic error becomes 500", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		RespondWithError(rec, req, fmt.Errorf("unexpected internal error"))

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("attaches request ID from context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Request-ID", "test-req-id-123")
		ctx := pkglog.WithRequest(req.Context(), req)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		RespondWithError(rec, req, NotFound(""))

		var body map[string]interface{}
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "test-req-id-123", body["request_id"])
	})

	t.Run("no request ID when context is empty", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		RespondWithError(rec, req, NotFound(""))

		var body map[string]interface{}
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		_, hasRequestID := body["request_id"]
		assert.False(t, hasRequestID, "request_id should be omitted when not set in context")
	})
}
