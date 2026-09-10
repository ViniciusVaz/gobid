package api

import (
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func (api *Api) BindRoutes() {
	isDev, err := strconv.ParseBool(os.Getenv("GOBID_CSRF_DEV_ENV"))
	if err != nil {
		panic("Missing enviroment configuration CSRF")
	}

	csrfMiddeware := csrf.Protect(
		[]byte(os.Getenv("GOBID_CSRF_KEY")),
		csrf.Secure(false), // apenas desenvolvimento
	)

	if isDev {
		api.Router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, csrf.PlaintextHTTPRequest(r))
			})
		})
	}

	api.Router.Use(csrfMiddeware)

	api.Router.Use(
		middleware.RequestID,
		middleware.Recoverer,
		middleware.Logger,
		api.Sessions.LoadAndSave,
	)

	api.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/csrfToken", api.HandleGetCSRFToken)
			r.Route("/users/", func(r chi.Router) {
				r.Post("/signup", api.handleSignupUser)
				r.Post("/login", api.handleLoginUser)
				r.Post("/logout", api.handleLogoutUser)
				r.With(api.AuthMiddleware).Post("/logout", api.handleLogoutUser)
			})

			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)
					r.Post("/", api.handleCreateProduct)

					r.Get("/ws/subscribe/{product_id}", api.handleSubscribeUserToAuction)
				})
			})
		})
	})
}
