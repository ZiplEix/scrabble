package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupReportRoutes(e *echo.Echo) {
	r := e.Group("/report", middleware.RequireAuth)

	r.GET("/:id", controller.GetReportByID)
	r.GET("", controller.GetAllReports)
	r.POST("", controller.CreateReport)

	r.PATCH("/:id", controller.UpdateReport)
	r.PUT("/:id/resolve", controller.ResolveReport)         // set the status of the report to "resolved"
	r.PUT("/:id/reject", controller.RejectReport)           // set the status of the report to "rejected"
	r.PUT("/:id/progress", controller.UpdateReportProgress) // set the status of the report to "in_progress"
	r.DELETE("/:id", controller.DeleteReport)

	r.GET("/me", controller.GetMyReports)
}
