package album

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leoluyi/go-api-template/internal/errors"
	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/leoluyi/go-api-template/pkg/pagination"
)

type resource struct {
	service Service
	logger  log.Logger
}

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r chi.Router, service Service, authHandler func(http.Handler) http.Handler, logger log.Logger) {
	rs := resource{service, logger}

	// public routes
	r.Get("/albums/{id}", rs.get)
	r.Get("/albums", rs.query)

	// protected routes — require a valid JWT
	r.Group(func(r chi.Router) {
		r.Use(authHandler)
		r.Post("/albums", rs.create)
		r.Put("/albums/{id}", rs.update)
		r.Delete("/albums/{id}", rs.delete)
	})
}

// @Summary      Get album
// @Description  Returns a single album by ID
// @Tags         Albums
// @Produce      json
// @Param        id   path      string  true  "Album ID"
// @Success      200  {object}  Album
// @Failure      404  {object}  errors.ErrorResponse
// @Failure      500  {object}  errors.ErrorResponse
// @Router       /albums/{id} [get]
func (rs resource) get(w http.ResponseWriter, r *http.Request) {
	album, err := rs.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		errors.RespondWithError(w, r, err)
		return
	}
	errors.RespondJSON(w, http.StatusOK, album)
}

// @Summary      List albums
// @Description  Returns a paginated list of albums
// @Tags         Albums
// @Produce      json
// @Param        page      query     int  false  "Page number"
// @Param        per_page  query     int  false  "Page size"
// @Success      200       {object}  pagination.Pages
// @Failure      500       {object}  errors.ErrorResponse
// @Router       /albums [get]
func (rs resource) query(w http.ResponseWriter, r *http.Request) {
	count, err := rs.service.Count(r.Context())
	if err != nil {
		errors.RespondWithError(w, r, err)
		return
	}
	pages := pagination.NewFromRequest(r, count)
	albums, err := rs.service.Query(r.Context(), pages.Offset(), pages.Limit())
	if err != nil {
		errors.RespondWithError(w, r, err)
		return
	}
	pages.Items = albums
	errors.RespondJSON(w, http.StatusOK, pages)
}

// @Summary      Create album
// @Description  Creates a new album
// @Tags         Albums
// @Accept       json
// @Produce      json
// @Param        album  body      CreateAlbumRequest  true  "Album to create"
// @Success      201    {object}  Album
// @Failure      400    {object}  errors.ErrorResponse
// @Failure      401    {object}  errors.ErrorResponse
// @Failure      500    {object}  errors.ErrorResponse
// @Security     BearerAuth
// @Router       /albums [post]
func (rs resource) create(w http.ResponseWriter, r *http.Request) {
	var input CreateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rs.logger.With(r.Context()).Error(err)
		errors.RespondWithError(w, r, errors.BadRequest(""))
		return
	}
	album, err := rs.service.Create(r.Context(), input)
	if err != nil {
		errors.RespondWithError(w, r, err)
		return
	}
	errors.RespondJSON(w, http.StatusCreated, album)
}

// @Summary      Update album
// @Description  Updates an existing album by ID
// @Tags         Albums
// @Accept       json
// @Produce      json
// @Param        id     path      string             true  "Album ID"
// @Param        album  body      UpdateAlbumRequest true  "Updated album data"
// @Success      200    {object}  Album
// @Failure      400    {object}  errors.ErrorResponse
// @Failure      401    {object}  errors.ErrorResponse
// @Failure      404    {object}  errors.ErrorResponse
// @Failure      500    {object}  errors.ErrorResponse
// @Security     BearerAuth
// @Router       /albums/{id} [put]
func (rs resource) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rs.logger.With(r.Context()).Error(err)
		errors.RespondWithError(w, r, errors.BadRequest(""))
		return
	}
	album, err := rs.service.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		errors.RespondWithError(w, r, err)
		return
	}
	errors.RespondJSON(w, http.StatusOK, album)
}

// @Summary      Delete album
// @Description  Deletes an album by ID
// @Tags         Albums
// @Produce      json
// @Param        id   path      string  true  "Album ID"
// @Success      200  {object}  Album
// @Failure      401  {object}  errors.ErrorResponse
// @Failure      404  {object}  errors.ErrorResponse
// @Failure      500  {object}  errors.ErrorResponse
// @Security     BearerAuth
// @Router       /albums/{id} [delete]
func (rs resource) delete(w http.ResponseWriter, r *http.Request) {
	album, err := rs.service.Delete(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		errors.RespondWithError(w, r, err)
		return
	}
	errors.RespondJSON(w, http.StatusOK, album)
}
