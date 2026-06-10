package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	// Disable all HTTP routes as SvelteKit now communicates directly with Supabase.
	// We only keep a health check route for monitoring.
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}

