package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupUsersRoutes(e *echo.Echo) {
	// authenticated routes under /users
	r := e.Group("/users", middleware.RequireAuth)
	r.GET("/suggest", controller.SuggestUsers)

	// public user info: /user/:id
	e.GET("/user/:id", controller.GetUserPublic)
}
