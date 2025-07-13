package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/go-chi/chi/v5"
)

func setupGameRoutes(r *chi.Mux) {
	r.Route("/game", func(r chi.Router) {
		r.With(middleware.RequireAuth).Post("/", controller.CreateGame)
		r.With(middleware.RequireAuth).Get("/{id}", controller.GetGame)
		r.With(middleware.RequireAuth).Post("/{id}/play", controller.PlayMove)
		r.With(middleware.RequireAuth).Get("/", controller.GetUserGames)
	})
}
