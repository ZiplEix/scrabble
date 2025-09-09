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
}
