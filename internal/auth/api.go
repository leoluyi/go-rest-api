package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leoluyi/go-api-template/internal/errors"
	"github.com/leoluyi/go-api-template/pkg/log"
)

// LoginRequest represents a login request body.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a successful login response.
type LoginResponse struct {
	Token string `json:"token"`
}

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(r chi.Router, service Service, logger log.Logger) {
	r.Post("/login", login(service, logger))
}

// login returns a handler that handles user login requests.
//
// @Summary      Authenticate user
// @Description  Authenticates a user and returns a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest   true  "Login credentials"
// @Success      200          {object}  LoginResponse
// @Failure      400          {object}  errors.ErrorResponse
// @Failure      401          {object}  errors.ErrorResponse
// @Router       /login [post]
func login(service Service, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.With(r.Context()).Errorf("invalid request: %v", err)
			errors.RespondWithError(w, r, errors.BadRequest(""))
			return
		}
		token, err := service.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			errors.RespondWithError(w, r, err)
			return
		}
		errors.RespondJSON(w, http.StatusOK, LoginResponse{Token: token})
	}
}
