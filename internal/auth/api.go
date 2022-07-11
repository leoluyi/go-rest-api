package auth

import (
	"net/http"

	"github.com/go-chi/chi"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// RegisterHandlers registers handlers for different HTTP requests.
// @Summary Authenticates a user
// @Description Authenticates a user and generates a JWT
// @Tags Auth
// @Produce json
// @Router /login/{id} [post]
// @Success 200
// @Failure 400
// @Failure 404
func RegisterHandlers(r chi.Router, service Service, logger log.Logger) {
	r.Post("/login", login(service, logger))
}

// login returns a handler that handles user login request.
func login(service Service, logger log.Logger) http.HandlerFunc {
	return func(c *routing.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.Read(&req); err != nil {
			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		token, err := service.Login(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			return err
		}
		return c.Write(struct {
			Token string `json:"token"`
		}{token})
	}
}
