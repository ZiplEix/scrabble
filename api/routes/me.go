package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func SetupMeRoutes(e *echo.Echo) {
	me := e.Group("/me", middleware.RequireAuth)

	me.GET("", controller.GetMe, middleware.RequireAuth)
	me.PUT("/prefs", controller.UpdatePrefs)
}
