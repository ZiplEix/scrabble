package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

var nonImplementedHandler = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func SetupRoutes(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the Scrabble API!"))
	})

	setupAuthRoutes(r)
	setupGameRoutes(r)
}
