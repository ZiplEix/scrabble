package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupNotificationsRoutes(e *echo.Echo) {
	r := e.Group("/notifications", middleware.RequireAuth)

	r.POST("/push-subscribe", controller.PushSubscribe)
	e.GET("/notifications/test", controller.SendTest)
}
