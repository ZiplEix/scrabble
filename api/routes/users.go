package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupUsersRoutes(e *echo.Echo) {
	r := e.Group("/users", middleware.RequireAuth)

	r.GET("/suggest", controller.SuggestUsers)
}
