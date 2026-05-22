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
	r.POST("/friends/:id", controller.AddFriend)
	r.DELETE("/friends/:id", controller.RemoveFriend)
	r.GET("/friends", controller.GetFriends)
	r.GET("/recent-opponents", controller.GetRecentOpponents)

	// public leaderboard
	e.GET("/leaderboard", controller.GetLeaderboard)
	e.GET("/stats/user/:id", controller.GetUserStats)

	// protected user info: /user/:id
	e.GET("/user/:id", controller.GetUserPublic, middleware.RequireAuth)
}
