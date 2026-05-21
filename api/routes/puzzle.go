package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupPuzzleRoutes(e *echo.Echo) {
	// Public routes (no auth required for leaderboard viewing)
	p := e.Group("/puzzles")

	// Protected routes (require auth)
	pAuth := e.Group("/puzzles", middleware.RequireAuth)

	// Get current puzzle of the day
	pAuth.GET("/today", controller.GetCurrentPuzzle)

	// Get specific puzzle by ID
	pAuth.GET("/:id", controller.GetPuzzleByID)

	// Get puzzle history for current user
	pAuth.GET("", controller.GetPuzzleHistory)

	// Submit an attempt for a puzzle
	pAuth.POST("/:id/attempts", controller.SubmitPuzzleAttempt)

	// Simulate score for current pending letters
	pAuth.POST("/:id/simulate_score", controller.SimulatePuzzleScore)

	// Start a puzzle (records server-side start timestamp)
	pAuth.POST("/:id/start", controller.StartPuzzle)

	// Get leaderboard for a puzzle (public)
	p.GET("/:id/leaderboard", controller.GetPuzzleLeaderboard)

	// Get player puzzle stats
	pAuth.GET("/me/stats", controller.GetPuzzleStats)

	// Admin routes
	admin := e.Group("/admin/puzzles", middleware.RequireAuth, middleware.RequireAdmin)
	admin.POST("/generate", controller.GeneratePuzzleAdmin)
}
