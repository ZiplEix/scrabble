package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupAdminRoutes(e *echo.Echo) {
	a := e.Group("/admin", middleware.RequireAuth, middleware.RequireAdmin)

	a.GET("/stats", controller.GetAdminStats)
	a.GET("/stats/logs", controller.GetLogsStats)
	a.GET("/logs/resume", controller.GetLogsResume)
	a.GET("/logs", controller.GetLogs)
	a.GET("/log/:id", controller.GetLogByID)
	a.GET("/users", controller.GetAllUsers)
	a.GET("/user/:id", controller.GetAdminUserByID)
	a.GET("/games", controller.GetAdminGames)
	a.GET("/game/:id", controller.GetAdminGameByID)
}
