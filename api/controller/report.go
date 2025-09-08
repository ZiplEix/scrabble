package controller

import (
	"fmt"
	"net/http"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

func CreateReport(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized, no user_id")
	}

	var req request.CreateReportRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.Title == "" || req.Content == "" {
		logctx.Add(c, "reason", "title_and_content_required")
		return echo.NewHTTPError(http.StatusBadRequest, "title and content are required")
	}

	reportID, err := services.CreateReport(userID, req.Title, req.Content)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_create_report",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to create report: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":   "Report created",
		"report_id": reportID,
	})
}

func GetReportByID(c echo.Context) error {
	reportID := c.Param("id")
	if reportID == "" {
		logctx.Add(c, "reason", "missing_report_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing report id")
	}

	report, err := services.GetReportByID(reportID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_fetch_report",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to fetch report: %v", err))
	}

	return c.JSON(http.StatusOK, report)
}

func GetAllReports(c echo.Context) error {
	reports, err := services.GetAllReports()
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_fetch_reports",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to fetch reports: %v", err))
	}

	return c.JSON(http.StatusOK, reports)
}

func ResolveReport(c echo.Context) error {
	reportID := c.Param("id")
	if reportID == "" {
		logctx.Add(c, "reason", "missing_report_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing report id")
	}

	if err := services.UpdateReportStatus(reportID, "resolved"); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_resolve_report",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to resolve report: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Report marked as resolved",
	})
}

func RejectReport(c echo.Context) error {
	reportID := c.Param("id")
	if reportID == "" {
		logctx.Add(c, "reason", "missing_report_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing report id")
	}

	if err := services.UpdateReportStatus(reportID, "rejected"); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_reject_report",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to reject report: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Report marked as rejected",
	})
}

func UpdateReportProgress(c echo.Context) error {
	reportID := c.Param("id")
	if reportID == "" {
		logctx.Add(c, "reason", "missing_report_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing report id")
	}

	if err := services.UpdateReportStatus(reportID, "in_progress"); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_mark_report_as_in_progress",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to mark report as in progress: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Report marked as in progress",
	})
}

func DeleteReport(c echo.Context) error {
	reportID := c.Param("id")
	if reportID == "" {
		logctx.Add(c, "reason", "missing_report_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing report id")
	}

	if err := services.DeleteReport(reportID); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_delete_report",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to delete report: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Report deleted",
	})
}

func UpdateReport(c echo.Context) error {
	reportID := c.Param("id")
	if reportID == "" {
		logctx.Add(c, "reason", "missing_report_id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing report id")
	}

	var req request.UpdateReportRequest
	if err := c.Bind(&req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "bind_failed",
			"body":   err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Title == "" && req.Content == "" && req.Status == "" {
		logctx.Add(c, "reason", "nothing_to_update")
		return echo.NewHTTPError(http.StatusBadRequest, "nothing to update")
	}

	if req.Status != "" && req.Status != "open" && req.Status != "in_progress" && req.Status != "resolved" && req.Status != "rejected" {
		logctx.Add(c, "reason", "invalid_status_value")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status value")
	}

	if err := services.UpdateReport(reportID, req); err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_update_report",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to update report: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Report updated",
	})
}

func GetMyReports(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		logctx.Add(c, "reason", "unauthorized")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	reports, err := services.GetReportsByUserID(userID)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_fetch_reports",
			"error":  err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to fetch reports: %v", err))
	}

	return c.JSON(http.StatusOK, reports)
}
