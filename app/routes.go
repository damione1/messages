package app

import (
	"log"
	"log/slog"
	"messages/app/handlers"
	"messages/app/views/errors"
	"messages/plugins/auth"
	"net/http"

	"github.com/anthdm/superkit/kit"
	"github.com/anthdm/superkit/kit/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/invopop/ctxi18n"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Define your global middleware
func InitializeMiddleware(router *chi.Mux) {
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(newLanguageMiddleware)
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

	router.Get("/set-language", HandleSetLanguage)

	// Routes that "might" have an authenticated user
	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, false)) // strict set to false

		// Routes
		app.Get("/api/messages", kit.Handler(handlers.HandleApi))
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

		app.Get("/users", kit.Handler(handlers.HandleUsersList))
		app.Delete("/user/{id}", kit.Handler(handlers.HandleUserDelete))

		app.Post("/invitation", kit.Handler(handlers.HandleInvitationCreate))
		app.Delete("/invitation/{id}", kit.Handler(handlers.HandleInvitationDelete))
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

func newLanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en" // Default language
		cookie, err := r.Cookie("lang")
		if err == nil {
			lang = cookie.Value
		}

		ctx, err := ctxi18n.WithLocale(r.Context(), lang)
		if err != nil {
			log.Printf("error setting locale: %v", err)
			http.Error(w, "error setting locale", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Define this in your auth or another appropriate handler file
func HandleSetLanguage(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		http.Error(w, "language query param is missing", http.StatusBadRequest)
		return
	}

	// Set the language cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "lang",
		Value: lang,
		Path:  "/",
		// Add any additional cookie properties, like Secure, HttpOnly, etc.
	})

	// Redirect back or handle accordingly
	http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect to home or a suitable location
}
