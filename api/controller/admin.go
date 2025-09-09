package controller

import (
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
