package album

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/leoluyi/go-api-template/internal/auth"
	"github.com/leoluyi/go-api-template/internal/entity"
	"github.com/leoluyi/go-api-template/internal/test"
	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/stretchr/testify/assert"
)

var errRepo = errors.New("repository error")

// errInjRepository is a Repository that returns configurable errors.
type errInjRepository struct {
	countErr  error
	queryErr  error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (r *errInjRepository) Get(_ context.Context, _ string) (entity.Album, error) {
	return entity.Album{}, r.getErr
}

func (r *errInjRepository) Count(_ context.Context) (int, error) {
	return 0, r.countErr
}

func (r *errInjRepository) Query(_ context.Context, _, _ int) ([]entity.Album, error) {
	return nil, r.queryErr
}

func (r *errInjRepository) Create(_ context.Context, _ entity.Album) error {
	return r.createErr
}

func (r *errInjRepository) Update(_ context.Context, _ entity.Album) error {
	return r.updateErr
}

func (r *errInjRepository) Delete(_ context.Context, _ string) error {
	return r.deleteErr
}

// TestAPIErrors exercises the HTTP 500 error paths in each handler.
func TestAPIErrors(t *testing.T) {
	logger, _ := log.NewForTest()
	header := auth.MockAuthHeader()

	t.Run("get — repo error returns 500", func(t *testing.T) {
		router := test.MockRouter(logger)
		RegisterHandlers(router, NewService(&errInjRepository{getErr: errRepo}, logger), auth.MockAuthHandler, logger)
		test.Endpoint(t, router, test.APITestCase{
			Name: "get repo error", Method: "GET", URL: "/albums/123",
			WantStatus: http.StatusInternalServerError,
		})
	})

	t.Run("query — Count error returns 500", func(t *testing.T) {
		router := test.MockRouter(logger)
		RegisterHandlers(router, NewService(&errInjRepository{countErr: errRepo}, logger), auth.MockAuthHandler, logger)
		test.Endpoint(t, router, test.APITestCase{
			Name: "query count error", Method: "GET", URL: "/albums",
			WantStatus: http.StatusInternalServerError,
		})
	})

	t.Run("query — Query error returns 500", func(t *testing.T) {
		// Count succeeds (nil error), but Query fails.
		router := test.MockRouter(logger)
		RegisterHandlers(router, NewService(&errInjRepository{queryErr: errRepo}, logger), auth.MockAuthHandler, logger)
		test.Endpoint(t, router, test.APITestCase{
			Name: "query list error", Method: "GET", URL: "/albums",
			WantStatus: http.StatusInternalServerError,
		})
	})

	t.Run("create — repo Create error returns 500", func(t *testing.T) {
		router := test.MockRouter(logger)
		RegisterHandlers(router, NewService(&errInjRepository{createErr: errRepo}, logger), auth.MockAuthHandler, logger)
		test.Endpoint(t, router, test.APITestCase{
			Name: "create repo error", Method: "POST", URL: "/albums",
			Body: `{"name":"test"}`, Header: header,
			WantStatus: http.StatusInternalServerError,
		})
	})

	t.Run("update — repo Get error returns 500", func(t *testing.T) {
		router := test.MockRouter(logger)
		RegisterHandlers(router, NewService(&errInjRepository{getErr: errRepo}, logger), auth.MockAuthHandler, logger)
		test.Endpoint(t, router, test.APITestCase{
			Name: "update repo error", Method: "PUT", URL: "/albums/123",
			Body: `{"name":"test"}`, Header: header,
			WantStatus: http.StatusInternalServerError,
		})
	})

	t.Run("update — not found returns 404", func(t *testing.T) {
		router := test.MockRouter(logger)
		RegisterHandlers(router, NewService(&errInjRepository{getErr: sql.ErrNoRows}, logger), auth.MockAuthHandler, logger)
		test.Endpoint(t, router, test.APITestCase{
			Name: "update not found", Method: "PUT", URL: "/albums/123",
			Body: `{"name":"test"}`, Header: header,
			WantStatus: http.StatusNotFound,
		})
	})

	t.Run("delete — repo error returns 500", func(t *testing.T) {
		router := test.MockRouter(logger)
		RegisterHandlers(router, NewService(&errInjRepository{getErr: errRepo}, logger), auth.MockAuthHandler, logger)
		test.Endpoint(t, router, test.APITestCase{
			Name: "delete repo error", Method: "DELETE", URL: "/albums/123",
			Header: header, WantStatus: http.StatusInternalServerError,
		})
	})
}

// TestServiceErrors verifies that service methods propagate repository errors.
func TestServiceErrors(t *testing.T) {
	logger, _ := log.NewForTest()
	ctx := context.Background()

	t.Run("Count propagates repo error", func(t *testing.T) {
		svc := NewService(&errInjRepository{countErr: errRepo}, logger)
		_, err := svc.Count(ctx)
		assert.Equal(t, errRepo, err)
	})

	t.Run("Query propagates repo error", func(t *testing.T) {
		svc := NewService(&errInjRepository{queryErr: errRepo}, logger)
		_, err := svc.Query(ctx, 0, 10)
		assert.Equal(t, errRepo, err)
	})

	t.Run("Get propagates repo error", func(t *testing.T) {
		svc := NewService(&errInjRepository{getErr: errRepo}, logger)
		_, err := svc.Get(ctx, "any-id")
		assert.Equal(t, errRepo, err)
	})

	t.Run("Delete propagates repo Delete error", func(t *testing.T) {
		// Get must succeed so the service proceeds to repo.Delete.
		repo := &errInjRepository{
			getErr:    nil,
			deleteErr: errRepo,
		}
		// Patch Get to return a real album.
		svc := NewService(&deleteFailRepository{}, logger)
		_, err := svc.Delete(ctx, "any-id")
		assert.Equal(t, errRepo, err)
		_ = repo // used above for documentation
	})

	t.Run("Create propagates repo Create error", func(t *testing.T) {
		svc := NewService(&errInjRepository{createErr: errRepo}, logger)
		_, err := svc.Create(ctx, CreateAlbumRequest{Name: "test"})
		assert.Equal(t, errRepo, err)
	})

	t.Run("Update propagates repo Update error", func(t *testing.T) {
		svc := NewService(&updateFailRepository{}, logger)
		_, err := svc.Update(ctx, "any-id", UpdateAlbumRequest{Name: "test"})
		assert.Equal(t, errRepo, err)
	})
}

// deleteFailRepository has Get succeed (so the service proceeds) but Delete fail.
type deleteFailRepository struct{}

func (r *deleteFailRepository) Get(_ context.Context, _ string) (entity.Album, error) {
	return entity.Album{ID: "any-id", Name: "test"}, nil
}

func (r *deleteFailRepository) Count(_ context.Context) (int, error)              { return 1, nil }
func (r *deleteFailRepository) Query(_ context.Context, _, _ int) ([]entity.Album, error) {
	return []entity.Album{{ID: "any-id", Name: "test"}}, nil
}
func (r *deleteFailRepository) Create(_ context.Context, _ entity.Album) error { return nil }
func (r *deleteFailRepository) Update(_ context.Context, _ entity.Album) error { return nil }
func (r *deleteFailRepository) Delete(_ context.Context, _ string) error        { return errRepo }

// updateFailRepository has Get succeed but Update fail.
type updateFailRepository struct{}

func (r *updateFailRepository) Get(_ context.Context, _ string) (entity.Album, error) {
	return entity.Album{ID: "any-id", Name: "test"}, nil
}

func (r *updateFailRepository) Count(_ context.Context) (int, error)              { return 1, nil }
func (r *updateFailRepository) Query(_ context.Context, _, _ int) ([]entity.Album, error) {
	return []entity.Album{{ID: "any-id", Name: "test"}}, nil
}
func (r *updateFailRepository) Create(_ context.Context, _ entity.Album) error { return nil }
func (r *updateFailRepository) Update(_ context.Context, _ entity.Album) error { return errRepo }
func (r *updateFailRepository) Delete(_ context.Context, _ string) error        { return nil }
