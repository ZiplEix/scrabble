package controller

import (
	"fmt"
	"strconv"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/labstack/echo/v4"
)

func GetAdminStats(c echo.Context) error {
	logctx.Add(c, "role", "admin")

	res, err := services.GetAdminStats()
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_admin_stats",
			"error":  err.Error(),
		})
		return c.JSON(500, echo.Map{
			"error":   "failed to get admin stats",
			"message": "Erreur lors de la récupération des statistiques administrateur",
		})
	}

	return c.JSON(200, res)
}

func GetLogsStats(c echo.Context) error {
	logctx.Add(c, "role", "admin")

	adminQ := c.QueryParam("admin")
	includeAdmin := adminQ == "true" || adminQ == "1"

	res, err := services.GetLogsStats(includeAdmin)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_logs_stats",
			"error":  err.Error(),
		})
		return c.JSON(500, echo.Map{
			"error":   "failed to get logs stats",
			"message": "Erreur lors de la récupération des statistiques de logs",
		})
	}

	return c.JSON(200, res)
}

func GetLogsResume(c echo.Context) error {
	logctx.Add(c, "role", "admin")
	adminQ := c.QueryParam("admin")
	includeAdmin := adminQ == "true" || adminQ == "1"
	res, err := services.GetLogsResume(10, includeAdmin)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_log_resume",
			"error":  err.Error(),
		})
		return c.JSON(500, echo.Map{
			"error":   "failed to get logs resume",
			"message": "Erreur lors de la récupération du résumé des logs",
		})
	}

	return c.JSON(200, res)
}

func GetLogs(c echo.Context) error {
	logctx.Add(c, "role", "admin")

	pageS := c.QueryParam("page")
	page, err := strconv.Atoi(pageS)
	if err != nil || page < 1 {
		page = 0
	}

	adminQ := c.QueryParam("admin")
	includeAdmin := adminQ == "true" || adminQ == "1"

	logs, err := services.GetLogs(page, includeAdmin)
	if err != nil {
		logctx.Merge(c, map[string]any{
			"reason": "failed_to_get_logs",
			"error":  fmt.Errorf("failed to get logs for page '%d': %w", page, err),
		})
		return c.JSON(500, echo.Map{
			"error":   "failed to get logs",
			"message": "Erreur lors de la récupération du résumé des logs",
		})
	}

	return c.JSON(200, echo.Map{
		"logs": logs,
	})
}

func GetLogByID(c echo.Context) error {
	logctx.Add(c, "role", "admin")
	idS := c.Param("id")
	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		logctx.Merge(c, map[string]any{"reason": "invalid_id", "id": idS, "error": err.Error()})
		return c.JSON(400, echo.Map{"error": "invalid id"})
	}

	adminQ := c.QueryParam("admin")
	includeAdmin := adminQ == "true" || adminQ == "1"

	entry, err := services.GetLogByID(id, includeAdmin)
	if err != nil {
		logctx.Merge(c, map[string]any{"reason": "failed_to_get_log", "error": err.Error()})
		return c.JSON(500, echo.Map{"error": "failed to get log"})
	}
	if entry == nil {
		return c.JSON(404, echo.Map{"error": "not found"})
	}

	return c.JSON(200, entry)
}
