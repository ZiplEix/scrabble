package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/go-chi/chi/v5"
)

func setupAuthRoutes(r *chi.Mux) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", controller.Register)
		r.Post("/login", controller.Login)
		r.With(middleware.RequireAuth).Get("/profile", nonImplementedHandler)
		// r.Get("/logout", nonImplementedHandler)
		// r.Get("/profile", nonImplementedHandler)
	})
}
