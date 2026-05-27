package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ZiplEix/scrabble/api/config"
	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/middleware/accesslog"
	requestid "github.com/ZiplEix/scrabble/api/middleware/request_id"
	"github.com/ZiplEix/scrabble/api/pkg/logger"
	"github.com/ZiplEix/scrabble/api/routes"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func init() {
	config.InitEnv()

	if err := database.Init(os.Getenv("POSTGRES_URL")); err != nil {
		panic(err)
	}

	if err := database.RunMigrations(embedMigrations); err != nil {
		panic(err)
	}
}

func initLogger() func(context.Context) error {
	mode := os.Getenv("LOG_MODE")
	if mode == "" {
		mode = "human" // beautiful terminal colors by default
	}

	pgClose := logger.Init(mode, database.DB)
	
	// Enable DB logging by default for the API
	logger.SaveToDB(true)

	return pgClose
}

func main() {
	pgClose := initLogger()

	// Start log retention: purge logs older than 7 days every 24h
	stopRetention := StartLogRetention(database.DB, 7*24*time.Hour, 24*time.Hour)

	// Initialiser le bot Scrabby et démarrer le worker de rattrapage
	services.InitBot()
	services.StartBotWorker(5) // poll toutes les 5 secondes

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = pgClose(ctx)
		_ = stopRetention(ctx)
	}()

	e := echo.New()

	e.Use(requestid.Middleware())
	e.Use(accesslog.Middleware())

	// e.Use(echozap.ZapLogger(zlog))
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())
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
