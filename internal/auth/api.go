package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// RegisterHandlers registers handlers for different HTTP requests.
// @Summary Authenticates a user
// @Description Authenticates a user and generates a JWT
// @Tags Auth
// @Produce json
// @Router /login [post]
// @Success 200
// @Failure 400
// @Failure 401
func RegisterHandlers(r chi.Router, service Service, logger log.Logger) {
	r.Post("/login", login(service, logger))
}

// login returns a handler that handles user login requests.
func login(service Service, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.With(r.Context()).Errorf("invalid request: %v", err)
			errors.RespondWithError(w, errors.BadRequest(""))
			return
		}
		token, err := service.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			errors.RespondWithError(w, err)
			return
		}
		errors.RespondJSON(w, http.StatusOK, struct {
			Token string `json:"token"`
		}{token})
	}
}
