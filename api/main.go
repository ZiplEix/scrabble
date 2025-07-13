package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ZiplEix/scrabble/api/config"
	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func init() {
	config.InitEnv()

	if err := database.Init(os.Getenv("POSTGRES_URL")); err != nil {
		panic(err)
	}

	database.Migrate()
}

func setupCors(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://scrabble.baptiste.zip"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // pré-cache pendant 5 min
	}))
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	setupCors(r)

	routes.SetupRoutes(r)

	fmt.Println("Server is running on https://localhost:8080")
	if err := http.ListenAndServe(":8888", r); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server stopped")
	os.Exit(0)
}
