package routes

import (
	"github.com/ZiplEix/scrabble/api/controller"
	"github.com/ZiplEix/scrabble/api/middleware"
	"github.com/labstack/echo/v4"
)

func setupReportRoutes(e *echo.Echo) {
	r := e.Group("/report", middleware.RequireAuth)

	r.POST("", controller.CreateReport)
	r.GET("/me", controller.GetMyReports)
	r.GET("/:id", controller.GetReportByID)

	r.GET("", controller.GetAllReports, middleware.RequireAdmin)
	r.PATCH("/:id", controller.UpdateReport, middleware.RequireAdmin)
	r.PUT("/:id/resolve", controller.ResolveReport, middleware.RequireAdmin)         // set the status of the report to "resolved"
	r.PUT("/:id/reject", controller.RejectReport, middleware.RequireAdmin)           // set the status of the report to "rejected"
	r.PUT("/:id/progress", controller.UpdateReportProgress, middleware.RequireAdmin) // set the status of the report to "in_progress"
	r.DELETE("/:id", controller.DeleteReport, middleware.RequireAdmin)
}
