package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ZiplEix/scrabble/api/config"
	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/middleware/accesslog"
	requestid "github.com/ZiplEix/scrabble/api/middleware/request_id"
	"github.com/ZiplEix/scrabble/api/pgzap"
	"github.com/ZiplEix/scrabble/api/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	config.InitEnv()

	if err := database.Init(os.Getenv("POSTGRES_URL")); err != nil {
		panic(err)
	}
}

func initLogger() (*zap.Logger, func(context.Context) error) {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	pgCore, pgClose := pgzap.New(
		database.DB,
		zapcore.InfoLevel,
		pgzap.WithBatchSize(1000),
		pgzap.WithMaxWait(5*time.Second),
		pgzap.WithBuffer(10_000),
	)

	core := zapcore.NewTee(consoleCore, pgCore)
	zlog := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(zlog)

	return zlog, pgClose
}

func main() {
	zlog, pgClose := initLogger()

	// Start log retention: purge logs older than 7 days every 24h
	stopRetention := StartLogRetention(database.DB, 7*24*time.Hour, 24*time.Hour)

	defer func() {
		_ = zlog.Sync()
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
