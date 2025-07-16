package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ZiplEix/scrabble/api/config"
	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	config.InitEnv()

	if err := database.Init(os.Getenv("POSTGRES_URL")); err != nil {
		panic(err)
	}

	if err := database.Migrate(); err != nil {
		panic(err)
	}
}

func main() {
	e := echo.New()

	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Bienvenue sur l'API")
	})

	routes.SetupRoutes(e)

	fmt.Println("Server is running on https://0.0.0.0:8888")
	if err := e.Start("0.0.0.0:8888"); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
