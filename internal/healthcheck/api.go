package healthcheck

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leoluyi/go-api-template/internal/errors"
)

// DBChecker is a minimal interface for checking database connectivity.
type DBChecker interface {
	PingContext(ctx context.Context) error
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	DB      string `json:"db"`
}

type handler struct {
	version string
	db      DBChecker
}

// RegisterHandlers registers the handlers that perform healthchecks.
func RegisterHandlers(r chi.Router, version string, db DBChecker) {
	h := handler{version: version, db: db}
	r.Get("/healthcheck", h.check)
	r.Head("/healthcheck", h.check)
}

// @Summary      Health check
// @Description  Returns the service status and database connectivity
// @Tags         System
// @Produce      json
// @Success      200  {object}  healthResponse
// @Failure      503  {object}  healthResponse
// @Router       /healthcheck [get]
func (h handler) check(w http.ResponseWriter, r *http.Request) {
	res := healthResponse{
		Status:  "ok",
		Version: h.version,
		DB:      "ok",
	}

	if h.db != nil {
		if err := h.db.PingContext(r.Context()); err != nil {
			res.Status = "degraded"
			res.DB = "error: " + err.Error()
			errors.RespondJSON(w, http.StatusServiceUnavailable, res)
			return
		}
	}

	errors.RespondJSON(w, http.StatusOK, res)
}
