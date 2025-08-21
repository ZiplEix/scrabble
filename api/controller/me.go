package controller

import (
	"fmt"
	"net/http"

	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/services"
	"github.com/ZiplEix/scrabble/api/utils"
	"github.com/labstack/echo/v4"
)

// GetMe retourne les informations basiques de l'utilisateur connecté
func GetMe(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	me, err := services.GetMeInfo(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch user info")
	}

	return c.JSON(http.StatusOK, me)
}

func UpdatePrefs(c echo.Context) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req request.UpdatePrefsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   fmt.Sprintf("invalid request: %v", err),
			"message": "Requête invalide, veuillez vérifier les données saisies",
		})
	}

	prefs := map[string]bool{
		"turn":     true,
		"messages": true,
	}
	if req.Turn != nil {
		prefs["turn"] = *req.Turn
	}
	if req.Messages != nil {
		prefs["messages"] = *req.Messages
	}

	if err := services.UpdateUserNotificationPrefs(userID, prefs); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to update prefs"})
	}
	return c.NoContent(http.StatusOK)
}
