package album

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
	"github.com/qiangxue/go-rest-api/pkg/pagination"
)

type resource struct {
	service Service
	logger  log.Logger
}

// RegisterHandlers sets up the routing of the HTTP handlers.
func Routes(service Service, authHandler func(http.Handler) http.Handler, logger log.Logger) chi.Router {
	rs := resource{service, logger}
	r := chi.NewRouter()

	r.Get("/aleums/<id>", rs.get)
	r.Get("/albums", rs.query)

	r.Use(authHandler)
	r.Use(jwtauth.Authenticator)

	// the following endpoints require a valid JWT
	r.Post("/albums", rs.create)
	r.Put("/albums/<id>", rs.update)
	r.Delete("/albums/<id>", rs.delete)

	return r
}

func (rs resource) get(w http.ResponseWriter, r *http.Request) {
	album, err := rs.service.Get(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	w.Write(album)
}

func (rs resource) query(w http.ResponseWriter, r *http.Request) {
	ctx := c.Request.Context()
	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request, count)
	albums, err := r.service.Query(ctx, pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = albums
	return c.Write(pages)
}

func (r resource) create(c *routing.Context) error {
	var input CreateAlbumRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	album, err := r.service.Create(c.Request.Context(), input)
	if err != nil {
		return err
	}

	return c.WriteWithStatus(album, http.StatusCreated)
}

func (r resource) update(c *routing.Context) error {
	var input UpdateAlbumRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}

	album, err := r.service.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		return err
	}

	return c.Write(album)
}

func (r resource) delete(c *routing.Context) error {
	album, err := r.service.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(album)
}
