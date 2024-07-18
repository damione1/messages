package main

import (
	"fmt"
	"log"
	"messages/app"
	"messages/app/locales"
	"messages/public"
	"net/http"
	"os"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
	"github.com/invopop/ctxi18n"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	if err := ctxi18n.LoadWithDefault(locales.LocalesFs, "en"); err != nil {
		log.Fatalf("error loading locales with default: %v", err)
	}

	router := chi.NewMux()

	app.InitializeMiddleware(router)

	if kit.IsDevelopment() {
		router.Handle("/public/*", disableCache(staticDev()))
	} else if kit.IsProduction() {
		router.Handle("/public/*", staticProd())
	}

	kit.UseErrorHandler(app.ErrorHandler)
	router.HandleFunc("/*", kit.Handler(app.NotFoundHandler))

	app.InitializeRoutes(router)
	app.RegisterEvents()

	listenAddr := os.Getenv("HTTP_LISTEN_ADDR")
	// In development link the full Templ proxy url.
	url := "http://localhost:7331"
	if kit.IsProduction() {
		url = fmt.Sprintf("http://localhost%s", listenAddr)
	}

	fmt.Printf("application running in %s at %s\n", kit.Env(), url)

	http.ListenAndServe(listenAddr, router)
}

func staticDev() http.Handler {
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public")))
}

func staticProd() http.Handler {
	return http.StripPrefix("/public/", http.FileServerFS(public.AssetsFS))
}

func disableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}
