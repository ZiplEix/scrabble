package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupGameRoutes(e *echo.Echo) {
	g := e.Group("/game", middleware.RequireAuth)

	g.POST("", controller.CreateGame)
	g.DELETE("/:id", controller.DeleteGame)
	g.GET("/:id", controller.GetGame)
	g.GET("/:id/messages", controller.GetMessages)
	g.GET("/:id/unread_messages_count", controller.GetUnreadCountForGameHandler)
	g.POST("/:id/messages/read", controller.MarkMessagesReadHandler)
	g.DELETE("/:id/messages/:msg_id", controller.DeleteMessage)
	g.POST("/:id/play", controller.PlayMove)
	g.GET("", controller.GetUserGames)
	g.PUT("/:id/rename", controller.RenameGame)
	g.GET("/:id/new_rack", controller.GetNewRack)
	g.POST("/:id/simulate_score", controller.SimulateScore)
	g.POST("/:id/message", controller.CreateMessage)
	g.POST("/:id/pass", controller.PassTurn)
}
