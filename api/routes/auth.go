package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupAuthRoutes(e *echo.Echo) {
	auth := e.Group("/auth")

	auth.POST("/register", controller.Register)
	auth.POST("/login", controller.Login)
	auth.GET("/profile", nonImplementedHandler, middleware.RequireAuth)
	auth.POST("/change-password", controller.ChangePassword, middleware.RequireAuth, middleware.RequireAdmin)
	auth.GET("/connect-as", controller.ConnectAS, middleware.RequireAuth, middleware.RequireAdmin)

	auth.POST("/admin/login", controller.AdminLogin)
}
