package album

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
	"github.com/qiangxue/go-rest-api/pkg/pagination"
)

type resource struct {
	service Service
	logger  log.Logger
}

// RegisterHandlers sets up the routing of the HTTP handlers.
// @Summary Album operations
// @Tags Albums
// @Produce json
// @Router /albums [get]
// @Router /albums/{id} [get]
// @Router /albums [post]
// @Router /albums/{id} [put]
// @Router /albums/{id} [delete]
func RegisterHandlers(r chi.Router, service Service, authHandler func(http.Handler) http.Handler, logger log.Logger) {
	rs := resource{service, logger}

	// public routes
	r.Get("/albums/{id}", rs.get)
	r.Get("/albums", rs.query)

	// protected routes — require a valid JWT
	r.Group(func(r chi.Router) {
		r.Use(authHandler)
		r.Use(jwtauth.Authenticator)
		r.Post("/albums", rs.create)
		r.Put("/albums/{id}", rs.update)
		r.Delete("/albums/{id}", rs.delete)
	})
}

func (rs resource) get(w http.ResponseWriter, r *http.Request) {
	album, err := rs.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		errors.RespondWithError(w, err)
		return
	}
	errors.RespondJSON(w, http.StatusOK, album)
}

func (rs resource) query(w http.ResponseWriter, r *http.Request) {
	count, err := rs.service.Count(r.Context())
	if err != nil {
		errors.RespondWithError(w, err)
		return
	}
	pages := pagination.NewFromRequest(r, count)
	albums, err := rs.service.Query(r.Context(), pages.Offset(), pages.Limit())
	if err != nil {
		errors.RespondWithError(w, err)
		return
	}
	pages.Items = albums
	errors.RespondJSON(w, http.StatusOK, pages)
}

func (rs resource) create(w http.ResponseWriter, r *http.Request) {
	var input CreateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rs.logger.With(r.Context()).Info(err)
		errors.RespondWithError(w, errors.BadRequest(""))
		return
	}
	album, err := rs.service.Create(r.Context(), input)
	if err != nil {
		errors.RespondWithError(w, err)
		return
	}
	errors.RespondJSON(w, http.StatusCreated, album)
}

func (rs resource) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rs.logger.With(r.Context()).Info(err)
		errors.RespondWithError(w, errors.BadRequest(""))
		return
	}
	album, err := rs.service.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		errors.RespondWithError(w, err)
		return
	}
	errors.RespondJSON(w, http.StatusOK, album)
}

func (rs resource) delete(w http.ResponseWriter, r *http.Request) {
	album, err := rs.service.Delete(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		errors.RespondWithError(w, err)
		return
	}
	errors.RespondJSON(w, http.StatusOK, album)
}
