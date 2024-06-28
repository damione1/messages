package app

import (
	"log/slog"
	"messages/app/handlers"
	"messages/app/views/errors"
	"messages/plugins/auth"

	"github.com/anthdm/superkit/kit"
	"github.com/anthdm/superkit/kit/middleware"
	"github.com/go-chi/chi/v5"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Define your global middleware
func InitializeMiddleware(router *chi.Mux) {
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.WithRequestURL)
}

// Define your routes in here
func InitializeRoutes(router *chi.Mux) {
	// Authentication plugin:
	auth.InitializeRoutes(router)

	authConfig := kit.AuthenticationConfig{
		AuthFunc:    auth.AuthenticateUser,
		RedirectURL: "/login",
	}

	// Routes that "might" have an authenticated user
	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, false)) // strict set to false

		// Routes
		app.Get("/api", kit.Handler(handlers.HandleApi))
	})

	// Authenticated routes
	//
	// Routes that "must" have an authenticated user or else they
	// will be redirected to the configured redirectURL, set in the
	// AuthenticationConfig.
	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, true)) // strict set to true

		app.Get("/", kit.Handler(func(kit *kit.Kit) error {
			return kit.Redirect(302, "/messages")
		}))

		app.Route("/message", func(r chi.Router) {
			r.Get("/{id}", kit.Handler(handlers.HandleMessageGet))
			r.Post("/", kit.Handler(handlers.HandleMessageCreate))
			r.Patch("/{id}", kit.Handler(handlers.HandleMessageUpdate))
			r.Delete("/{id}", kit.Handler(handlers.HandleMessageDelete))

			r.Get("/", kit.Handler(func(kit *kit.Kit) error {
				return kit.Redirect(302, "/messages")
			}))
		})
		app.Get("/messages", kit.Handler(handlers.HandleMessagesList))

		app.Route("/website", func(r chi.Router) {
			r.Get("/{id}", kit.Handler(handlers.HandleWebsiteGet))
			r.Post("/", kit.Handler(handlers.HandleWebsiteCreate))
			r.Patch("/{id}", kit.Handler(handlers.HandleWebsiteUpdate))
			r.Delete("/{id}", kit.Handler(handlers.HandleWebsiteDelete))

			r.Get("/", kit.Handler(func(kit *kit.Kit) error {
				return kit.Redirect(302, "/websites")
			}))
		})
		app.Get("/websites", kit.Handler(handlers.HandleWebsitesList))
	})
}

// NotFoundHandler that will be called when the requested path could
// not be found.
func NotFoundHandler(kit *kit.Kit) error {
	return kit.Render(errors.Error404())
}

// ErrorHandler that will be called on errors return from application handlers.
func ErrorHandler(kit *kit.Kit, err error) {
	slog.Error("internal server error", "err", err.Error(), "path", kit.Request.URL.Path)
	kit.Render(errors.Error500())
}
