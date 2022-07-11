package healthcheck

import (
	"net/http"

	"github.com/go-chi/chi"
)

// RegisterHandlers registers the handlers that perform healthchecks.
func RegisterHandlers(r chi.Router, version string) {
	r.Get("/healthcheck", healthcheck(version))
	r.Head("/healthcheck", healthcheck(version))
}

// healthcheck responds to a healthcheck request.
func healthcheck(version string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK " + version))
	}
}
