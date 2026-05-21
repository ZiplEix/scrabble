package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/labstack/echo/v4"
)

func setupDictionaryRoutes(e *echo.Echo) {
	e.GET("/dictionary/:word", controller.GetDictionaryDefinition)
	e.POST("/dictionary", controller.SaveDictionaryDefinition)
}
