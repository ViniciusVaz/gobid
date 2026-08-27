package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (api *Api) BindRoutes() {

	api.Router.Route("/api", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Route("/v1", func(r chi.Router) {
			r.Route("/users/", func(r chi.Router) {
				r.Post("/signup", api.handleSignupUser)
				r.Post("/login", api.handleLoginUser)
				r.Post("/logout", api.handleLogoutUser)
			})
		})
	})
}
